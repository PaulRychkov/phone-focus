package dev.rychkov.phonefocus

import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent

class BootReceiver : BroadcastReceiver() {
    override fun onReceive(context: Context, intent: Intent?) {
        val prefs = context.getSharedPreferences(SamplerService.PREFS, Context.MODE_PRIVATE)
        if (!prefs.getBoolean("enabled", false)) return
        runCatching { SamplerService.start(context.applicationContext) }
    }
}
