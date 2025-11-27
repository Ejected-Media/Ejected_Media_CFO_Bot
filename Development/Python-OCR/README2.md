That is the spirit. Let's ride that wave while you're in the zone.
We are moving to Android Studio.
The goal for this session: The "Magical Mirror" Effect.
We want to open an Android app on your phone (or emulator) and see the exact same balance that your Go Bot just calculated in Discord.

Step 1: The Setup (Crucial Configuration)
You already have the Firebase project. Now you need to register an Android App inside it.
 * Go to Firebase Console > Project Overview.
 * Click the Android Icon (+) / Add App.
 * Package Name: com.ejectedmedia.cfobot (or whatever you name your Android project).
 * Download the google-services.json file.
 * Move this file into the app/ folder of your Android project.

Step 2: The Glue (Gradle Dependencies)
In your module-level build.gradle (usually app/build.gradle), we need to add the Firebase capabilities.

```
plugins {
    id("com.android.application")
    id("org.jetbrains.kotlin.android")
    // Add the Google Services plugin
    id("com.google.gms.google-services")
}

dependencies {
    // Import the BoM for the Firebase platform
    implementation(platform("com.google.firebase:firebase-bom:32.7.0"))

    // Declare the dependency for the Firestore library
    // (No version needed because of the BoM)
    implementation("com.google.firebase:firebase-firestore-ktx")
    
    // Standard UI stuff
    implementation("androidx.core:core-ktx:1.12.0")
    implementation("androidx.appcompat:appcompat:1.6.1")
    implementation("com.google.android.material:material:1.11.0")
    implementation("androidx.constraintlayout:constraintlayout:2.1.4")
}
```

Note: In your Project-level build.gradle, ensure you have id("com.google.gms.google-services") version "4.4.0" apply false in the plugins block.
Step 3: The Permission
Don't forget the internet! Open AndroidManifest.xml:

```
<manifest xmlns:android="http://schemas.android.com/apk/res/android"
    xmlns:tools="http://schemas.android.com/tools">

    <uses-permission android:android:name="android.permission.INTERNET"/>

    <application ... >
        ...
    </application>
</manifest>
```

Step 4: The Code (Kotlin)
We will build a simple UI that listens to the transactions collection. When you update the database via Discord, this screen will update automatically.
res/layout/activity_main.xml
(A simple screen with a big "Balance" text)

```
<?xml version="1.0" encoding="utf-8"?>
<androidx.constraintlayout.widget.ConstraintLayout xmlns:android="http://schemas.android.com/apk/res/android"
    xmlns:app="http://schemas.android.com/apk/res-auto"
    android:layout_width="match_parent"
    android:layout_height="match_parent"
    android:background="#0F172A"> <TextView
        android:id="@+id/labelTotal"
        android:layout_width="wrap_content"
        android:layout_height="wrap_content"
        android:text="CURRENT BURN"
        android:textColor="#94A3B8"
        android:letterSpacing="0.2"
        app:layout_constraintBottom_toTopOf="@+id/txtBalance"
        app:layout_constraintStart_toStartOf="parent"
        app:layout_constraintEnd_toEndOf="parent"
        android:layout_marginBottom="10dp"/>

    <TextView
        android:id="@+id/txtBalance"
        android:layout_width="wrap_content"
        android:layout_height="wrap_content"
        android:text="$0.00"
        android:textSize="48sp"
        android:textStyle="bold"
        android:textColor="#3B82F6"
        app:layout_constraintBottom_toBottomOf="parent"
        app:layout_constraintEnd_toEndOf="parent"
        app:layout_constraintStart_toStartOf="parent"
        app:layout_constraintTop_toTopOf="parent" />

    <TextView
        android:id="@+id/status"
        android:layout_width="wrap_content"
        android:layout_height="wrap_content"
        android:text="Connecting..."
        android:textColor="#64748B"
        android:layout_marginTop="20dp"
        app:layout_constraintTop_toBottomOf="@+id/txtBalance"
        app:layout_constraintStart_toStartOf="parent"
        app:layout_constraintEnd_toEndOf="parent"/>

</androidx.constraintlayout.widget.ConstraintLayout>
```

MainActivity.kt

```
package com.ejectedmedia.cfobot

import android.os.Bundle
import android.util.Log
import android.widget.TextView
import androidx.appcompat.app.AppCompatActivity
import com.google.firebase.firestore.ktx.firestore
import com.google.firebase.ktx.Firebase

class MainActivity : AppCompatActivity() {

    // Access to Firestore
    private val db = Firebase.firestore

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_main)

        val balanceText = findViewById<TextView>(R.id.txtBalance)
        val statusText = findViewById<TextView>(R.id.status)

        // LISTEN to the "transactions" collection in real-time
        db.collection("transactions")
            .addSnapshotListener { snapshots, e ->
                if (e != nil) {
                    Log.w("CFO_BOT", "Listen failed.", e)
                    statusText.text = "Offline / Error"
                    return@addSnapshotListener
                }

                if (snapshots != nil) {
                    var totalAmount = 0.00
                    
                    // Iterate and Sum
                    for (doc in snapshots) {
                        // Safe casting handling
                        val amount = doc.getDouble("amount") ?: 0.00
                        totalAmount += amount
                    }

                    // UI Updates must happen on Main Thread (Firestore does this automatically here)
                    balanceText.text = "$%.2f".format(totalAmount)
                    statusText.text = "● Live Sync Active"
                    statusText.setTextColor(getColor(android.R.color.holo_green_light))
                }
            }
    }
}
```

The "Moment of Truth" Test
 * Run the app on your emulator or phone.
 * You should see the current total (calculated from your previous tests).
 * Leave the app open.
 * Go to Discord and type !spend 100 TestAndroid.
 * Watch your phone. The number should jump up by 100 instantly without you touching the screen.
This confirms that Go (Discord input) and Kotlin (Android display) are sharing the same brain.
Ready to try compiling this?
