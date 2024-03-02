package com.sashankg.hold

import android.content.Context
import android.provider.MediaStore
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
import retrofit2.Call
import retrofit2.http.POST
import java.util.concurrent.TimeUnit

@HiltWorker
class BackupWorker @AssistedInject constructor(
    @Assisted appContext: Context,
    @Assisted workerParams: WorkerParameters,
    private val mediaDao: MediaDao
) :
    CoroutineWorker(appContext, workerParams) {

    override suspend fun doWork(): Result {
        println(triggeredContentUris)
        return Result.success()
    }

    companion object {
        fun enqueue(context: Context) {
            val constraints = Constraints.Builder()
                .build()
            val work = OneTimeWorkRequestBuilder<WatcherWorker>().setConstraints(constraints).build()
            WorkManager.getInstance(context)
                .enqueueUniqueWork("backupworker", ExistingWorkPolicy.REPLACE, work)
        }
    }
}


interface ServerService {
    @POST("upload")
    fun upload(): Call<String>
}