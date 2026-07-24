package dev.rychkov.phonefocus

import android.app.Activity
import android.content.Context
import android.content.Intent
import android.net.Uri
import android.os.Build
import android.os.Bundle
import android.os.PowerManager
import android.provider.Settings
import android.webkit.JavascriptInterface
import android.webkit.WebView
import focus.Focus
import org.json.JSONObject

class MainActivity : Activity() {

    private lateinit var web: WebView

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        SamplerService.configure(applicationContext)

        val html = assets.open("index.html").bufferedReader().use { it.readText() }
        web = WebView(this).apply {
            settings.javaScriptEnabled = true
            settings.domStorageEnabled = true
            addJavascriptInterface(Bridge(), "Bridge")
            loadDataWithBaseURL("file:///android_asset/", html, "text/html", "utf-8", null)
        }
        setContentView(web)

        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
            requestPermissions(arrayOf(android.Manifest.permission.POST_NOTIFICATIONS), 1)
        }
    }

    override fun onResume() {
        super.onResume()
        if (this::web.isInitialized) {
            web.evaluateJavascript("window.refresh && window.refresh()", null)
        }
    }

    private fun prefs() = getSharedPreferences(SamplerService.PREFS, Context.MODE_PRIVATE)

    private fun open(intent: Intent) {
        runCatching { startActivity(intent.addFlags(Intent.FLAG_ACTIVITY_NEW_TASK)) }
    }

    private inner class Bridge {

        @JavascriptInterface
        fun status(): String {
            val power = getSystemService(Context.POWER_SERVICE) as PowerManager
            val saved = prefs()
            val version = runCatching {
                packageManager.getPackageInfo(packageName, 0).versionName
            }.getOrNull() ?: "?"
            return JSONObject()
                .put("version", version)
                .put("usageAccess", SamplerService.hasAccess(this@MainActivity))
                .put("batteryFree", power.isIgnoringBatteryOptimizations(packageName))
                .put("enabled", saved.getBoolean("enabled", false))
                .put("url", saved.getString("url", ""))
                .put("token", saved.getString("token", ""))
                .put("chat", saved.getString("chat", ""))
                .put("summary", Focus.summary())
                .put("pending", Focus.pending())
                .toString()
        }

        @JavascriptInterface
        fun snapshot(): String = Focus.lastSnapshot()

        @JavascriptInterface
        fun sampleNow() {
            SamplerService.configure(applicationContext)
            SamplerService.sampleAsync(applicationContext)
        }

        @JavascriptInterface
        fun save(url: String, token: String, chat: String) {
            prefs().edit()
                .putString("url", url.trim())
                .putString("token", token.trim())
                .putString("chat", chat.trim())
                .putBoolean("enabled", true)
                .apply()
            SamplerService.configure(applicationContext)
            SamplerService.start(applicationContext)
        }

        @JavascriptInterface
        fun stop() {
            prefs().edit().putBoolean("enabled", false).apply()
            SamplerService.stop(applicationContext)
        }

        @JavascriptInterface
        fun openUsageAccess() = open(Intent(Settings.ACTION_USAGE_ACCESS_SETTINGS))

        @JavascriptInterface
        fun openBattery() = open(
            Intent(Settings.ACTION_REQUEST_IGNORE_BATTERY_OPTIMIZATIONS)
                .setData(Uri.parse("package:$packageName"))
        )
    }
}
