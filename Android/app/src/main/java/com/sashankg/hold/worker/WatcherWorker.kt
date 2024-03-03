package com.sashankg.hold.worker

import android.content.Context
import android.provider.MediaStore
import androidx.hilt.work.HiltWorker
import androidx.work.Constraints
import androidx.work.CoroutineWorker
import androidx.work.ExistingWorkPolicy
import androidx.work.OneTimeWorkRequestBuilder
import androidx.work.WorkManager
import androidx.work.WorkerParameters
import com.sashankg.hold.model.Media
import com.sashankg.hold.model.MediaDao
import dagger.assisted.Assisted
import dagger.assisted.AssistedInject
import java.util.Date

@HiltWorker
class WatcherWorker @AssistedInject constructor(
    @Assisted context: Context,
    @Assisted workerParameters: WorkerParameters,
    private val mediaDao: MediaDao
) : CoroutineWorker(context, workerParameters) {

    override suspend fun doWork(): Result {
        val mediaList = mutableListOf<Media>()
        applicationContext.contentResolver.query(
            MediaStore.Images.Media.EXTERNAL_CONTENT_URI,
            null,
            null,
            null,
            null,
        )?.use { cursor ->
            val idColumn = cursor.getColumnIndexOrThrow(MediaStore.Video.Media._ID)
            while (cursor.moveToNext()) {
                val media = Media(cursor.getLong(idColumn), Date(), null)
                mediaList.add(media)
            }
        }
        println(mediaList)
        mediaDao.insertAll(mediaList)

        return Result.success()
    }

    companion object {
        fun enqueue(context: Context) {
            val work = OneTimeWorkRequestBuilder<WatcherWorker>().build()
            WorkManager.getInstance(context)
                .enqueueUniqueWork("backupworker", ExistingWorkPolicy.REPLACE, work)
        }
    }
}