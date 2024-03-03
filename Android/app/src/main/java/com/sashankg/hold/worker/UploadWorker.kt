package com.sashankg.hold.worker

import android.content.ContentUris
import android.content.Context
import android.net.Uri
import android.provider.MediaStore
import androidx.hilt.work.HiltWorker
import androidx.work.CoroutineWorker
import androidx.work.Data
import androidx.work.WorkerParameters
import com.sashankg.hold.ServerService
import com.sashankg.hold.model.MediaDao
import dagger.assisted.Assisted
import dagger.assisted.AssistedInject
import okhttp3.MediaType
import okhttp3.MultipartBody
import okhttp3.RequestBody
import okio.BufferedSink
import okio.ByteString
import retrofit2.await

@HiltWorker
class UploadWorker @AssistedInject constructor(
    @Assisted appContext: Context,
    @Assisted workerParams: WorkerParameters,
    private val mediaDao: MediaDao,
    private val server: ServerService
) : CoroutineWorker(appContext, workerParams) {
    override suspend fun doWork(): Result {
        val id = inputData.getLong("id", -1)
        if (id < 0) {
            Result.failure()
        }
        val contentUri: Uri = ContentUris.withAppendedId(
            MediaStore.Images.Media.EXTERNAL_CONTENT_URI,
            id
        )
        val call = server.upload(
            MultipartBody.Builder().addFormDataPart("file", "${contentUri}.jpg", object : RequestBody() {
                override fun contentType(): MediaType? {
                    return MediaType.parse("image/jpg")
                }

                override fun writeTo(sink: BufferedSink) {
                    applicationContext.contentResolver.openInputStream(contentUri)
                        ?.use { inputStream ->
                            val buf = ByteArray(DEFAULT_BUFFER_SIZE)
                            while (inputStream.read(buf) > 0) {
                                sink.outputStream().write(buf)
                            }
                            sink.write(ByteString.of(0))
                        }
                }
            }).build()
        ).await()

        println(call)

        return Result.success()
    }

    companion object {
        fun buildData(id: Long): Data {
            return Data.Builder()
                .putLong("id", id)
                .build()
        }
    }
}

