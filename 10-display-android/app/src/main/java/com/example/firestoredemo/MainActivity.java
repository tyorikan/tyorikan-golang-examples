package com.example.firestoredemo;

import android.os.Bundle;
import android.util.Log;

import androidx.appcompat.app.AppCompatActivity;

import com.google.firebase.firestore.FirebaseFirestore;
import com.google.firebase.firestore.QueryDocumentSnapshot;

public class MainActivity extends AppCompatActivity {
    private static final String TAG = MainActivity.class.toString();

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_main);

        FirebaseFirestore db = FirebaseFirestore.getInstance();
        db.collection("plate-states-160").get().addOnCompleteListener(task -> {
            if (!task.isSuccessful()) {
                Log.w(TAG, "Error getting documents.", task.getException());
                return;
            }

            for (QueryDocumentSnapshot document : task.getResult()) {
                Log.d(TAG, document.getId() + " => " + document.getData());

                db.collection("plate-states-160")
                        .document(document.getId())
                        .addSnapshotListener((value, error) -> Log.d(TAG, value.getId() + " changed to => " + value.getData()));
            }
        });
    }
}