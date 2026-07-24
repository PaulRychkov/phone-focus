package dev.rychkov.phonefocus

import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent
import kotlin.concurrent.thread

class SampleReceiver : BroadcastReceiver() {
    override fun onReceive(context: Context, intent: Intent?) {
        val app = context.applicationContext
        if (!app.getSharedPreferences(SamplerService.PREFS, Context.MODE_PRIVATE)
                .getBoolean("enabled", false)
        ) return
        val pending = goAsync()
        thread {
            try {
                SamplerService.configure(app)
                SamplerService.sampleNow(app)
                SamplerService.scheduleNext(app)
                runCatching { SamplerService.start(app) }
            } finally {
                pending.finish()
            }
        }
    }
}
