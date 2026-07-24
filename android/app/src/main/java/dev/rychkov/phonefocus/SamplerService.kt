package dev.rychkov.phonefocus

import android.app.AlarmManager
import android.app.AppOpsManager
import android.app.Notification
import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.PendingIntent
import android.app.Service
import android.app.usage.UsageEvents
import android.app.usage.UsageStatsManager
import android.content.Context
import android.content.Intent
import android.content.pm.ApplicationInfo
import android.content.pm.PackageManager
import android.content.pm.ServiceInfo
import android.os.Build
import android.os.IBinder
import android.os.Process
import focus.Focus
import org.json.JSONObject
import java.util.Calendar
import kotlin.concurrent.thread

class SamplerService : Service() {

    override fun onBind(intent: Intent?): IBinder? = null

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        configure(applicationContext)
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.UPSIDE_DOWN_CAKE) {
            startForeground(NOTIFICATION_ID, notification(), ServiceInfo.FOREGROUND_SERVICE_TYPE_SPECIAL_USE)
        } else {
            startForeground(NOTIFICATION_ID, notification())
        }
        sampleAsync(applicationContext)
        scheduleNext(applicationContext)
        return START_STICKY
    }

    private fun notification(): Notification {
        val manager = getSystemService(NotificationManager::class.java)
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O && manager.getNotificationChannel(CHANNEL) == null) {
            manager.createNotificationChannel(
                NotificationChannel(CHANNEL, getString(R.string.channel_name), NotificationManager.IMPORTANCE_LOW)
            )
        }
        val open = PendingIntent.getActivity(
            this, 0, Intent(this, MainActivity::class.java),
            PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_IMMUTABLE
        )
        val builder = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            Notification.Builder(this, CHANNEL)
        } else {
            @Suppress("DEPRECATION")
            Notification.Builder(this).setPriority(Notification.PRIORITY_LOW)
        }
        return builder
            .setContentTitle(getString(R.string.notif_title))
            .setContentText(getString(R.string.notif_text))
            .setSmallIcon(R.drawable.ic_stat)
            .setOngoing(true)
            .setContentIntent(open)
            .build()
    }

    companion object {
        const val PREFS = "phone_focus"
        const val INTERVAL_MS = 10L * 60 * 1000
        const val WINDOW_MS = 10L * 60 * 1000
        const val ACTION_SAMPLE = "dev.rychkov.phonefocus.SAMPLE"
        private const val NOTIFICATION_ID = 1001
        private const val CHANNEL = "sampling"

        fun sampleAsync(context: Context) {
            val app = context.applicationContext
            thread { sampleNow(app) }
        }

        fun sampleNow(context: Context) {
            if (!hasAccess(context)) return
            val now = System.currentTimeMillis()
            val dayStart = startOfDay(now)
            val raw = readEvents(context, dayStart, now)
            val appInfo = appInfoJson(context, raw)
            runCatching { Focus.push(raw, appInfo, WINDOW_MS / 1000, now) }
        }

        private fun startOfDay(now: Long): Long {
            val cal = Calendar.getInstance()
            cal.timeInMillis = now
            cal.set(Calendar.HOUR_OF_DAY, 0)
            cal.set(Calendar.MINUTE, 0)
            cal.set(Calendar.SECOND, 0)
            cal.set(Calendar.MILLISECOND, 0)
            return cal.timeInMillis
        }

        fun readEvents(context: Context, since: Long, until: Long): String {
            val manager = context.getSystemService(Context.USAGE_STATS_SERVICE) as UsageStatsManager
            val builder = StringBuilder()
            val stream = manager.queryEvents(since, until)
            val event = UsageEvents.Event()
            while (stream.hasNextEvent()) {
                stream.getNextEvent(event)
                builder.append(event.timeStamp).append('\t')
                    .append(event.eventType).append('\t')
                    .append(event.packageName ?: "").append('\n')
            }
            return builder.toString()
        }

        private fun appInfoJson(context: Context, raw: String): String {
            val pm = context.packageManager
            val packages = raw.lineSequence()
                .mapNotNull { it.split('\t').getOrNull(2)?.trim() }
                .filter { it.isNotEmpty() }
                .toSet()
            val root = JSONObject()
            for (pkg in packages) {
                try {
                    val info = pm.getApplicationInfo(pkg, 0)
                    val entry = JSONObject()
                        .put("label", pm.getApplicationLabel(info).toString())
                        .put("category", categoryId(info))
                    root.put(pkg, entry)
                } catch (e: PackageManager.NameNotFoundException) {
                }
            }
            return root.toString()
        }

        private fun categoryId(info: ApplicationInfo): String {
            if (Build.VERSION.SDK_INT < Build.VERSION_CODES.O) return ""
            return when (info.category) {
                ApplicationInfo.CATEGORY_GAME -> "games"
                ApplicationInfo.CATEGORY_SOCIAL -> "social"
                ApplicationInfo.CATEGORY_VIDEO -> "video"
                ApplicationInfo.CATEGORY_NEWS -> "reading"
                ApplicationInfo.CATEGORY_PRODUCTIVITY -> "productivity"
                else -> ""
            }
        }

        fun hasAccess(context: Context): Boolean {
            val ops = context.getSystemService(Context.APP_OPS_SERVICE) as AppOpsManager
            val mode = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.Q) {
                ops.unsafeCheckOpNoThrow(AppOpsManager.OPSTR_GET_USAGE_STATS, Process.myUid(), context.packageName)
            } else {
                @Suppress("DEPRECATION")
                ops.checkOpNoThrow(AppOpsManager.OPSTR_GET_USAGE_STATS, Process.myUid(), context.packageName)
            }
            if (mode == AppOpsManager.MODE_DEFAULT) {
                return context.checkCallingOrSelfPermission(
                    android.Manifest.permission.PACKAGE_USAGE_STATS
                ) == PackageManager.PERMISSION_GRANTED
            }
            return mode == AppOpsManager.MODE_ALLOWED
        }

        fun configure(context: Context) {
            val prefs = context.getSharedPreferences(PREFS, Context.MODE_PRIVATE)
            runCatching {
                Focus.configure(
                    prefs.getString("url", "") ?: "",
                    prefs.getString("token", "") ?: "",
                    Build.MODEL ?: "phone",
                    prefs.getString("chat", "")?.toLongOrNull() ?: 0L,
                    context.filesDir.absolutePath
                )
            }
        }

        private fun alarmPending(context: Context): PendingIntent {
            val intent = Intent(context, SampleReceiver::class.java).setAction(ACTION_SAMPLE)
            return PendingIntent.getBroadcast(
                context, 0, intent,
                PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_IMMUTABLE
            )
        }

        fun scheduleNext(context: Context) {
            val am = context.getSystemService(Context.ALARM_SERVICE) as AlarmManager
            val at = System.currentTimeMillis() + INTERVAL_MS
            try {
                am.setAndAllowWhileIdle(AlarmManager.RTC_WAKEUP, at, alarmPending(context))
            } catch (e: Exception) {
            }
        }

        fun cancel(context: Context) {
            (context.getSystemService(Context.ALARM_SERVICE) as AlarmManager).cancel(alarmPending(context))
        }

        fun start(context: Context) {
            val intent = Intent(context, SamplerService::class.java)
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
                context.startForegroundService(intent)
            } else {
                context.startService(intent)
            }
        }

        fun stop(context: Context) {
            cancel(context)
            context.stopService(Intent(context, SamplerService::class.java))
        }
    }
}
