package com.sashankg.hold.worker

import android.content.Context
import androidx.hilt.work.HiltWorker
import androidx.work.Constraints
import androidx.work.CoroutineWorker
import androidx.work.ExistingWorkPolicy
import androidx.work.OneTimeWorkRequestBuilder
import androidx.work.WorkManager
import androidx.work.WorkerParameters
import com.sashankg.hold.model.MediaDao
import dagger.assisted.Assisted
import dagger.assisted.AssistedInject

@HiltWorker
class BackupWorker @AssistedInject constructor(
    @Assisted appContext: Context,
    @Assisted workerParams: WorkerParameters,
    private val mediaDao: MediaDao
) :
    CoroutineWorker(appContext, workerParams) {

    override suspend fun doWork(): Result {
        val media = mediaDao.getAllMedia()
        WorkManager.getInstance(applicationContext).enqueue(
            media.map { media ->
                OneTimeWorkRequestBuilder<UploadWorker>()
                    .setInputData(UploadWorker.buildData(media.id))
                    .build()
            }
        )
        return Result.success()
    }

    companion object {
        fun enqueue(context: Context) {
            val constraints = Constraints.Builder()
                .build()
            val work = OneTimeWorkRequestBuilder<BackupWorker>().setConstraints(constraints).build()
            WorkManager.getInstance(context)
                .enqueueUniqueWork("backupworker", ExistingWorkPolicy.REPLACE, work)
        }
    }
}