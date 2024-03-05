package com.sashankg.hold

import dagger.Module
import dagger.Provides
import dagger.hilt.InstallIn
import dagger.hilt.components.SingletonComponent
import okhttp3.MultipartBody
import okhttp3.OkHttpClient
import okhttp3.RequestBody
import okhttp3.ResponseBody
import retrofit2.Call
import retrofit2.Response
import retrofit2.Retrofit
import retrofit2.http.Headers
import retrofit2.http.Multipart
import retrofit2.http.POST
import retrofit2.http.Part
import retrofit2.converter.scalars.ScalarsConverterFactory
import java.time.Duration
import java.util.concurrent.TimeUnit

interface ServerService {
    @Multipart
    @Headers("Connection: keep-alive")
    @POST("/upload")
    fun upload(@Part file: MultipartBody.Part): Call<String>
}

@Module
@InstallIn(SingletonComponent::class)
object ServerModule {


    val rf: Retrofit


    init {
        val client = OkHttpClient.Builder().writeTimeout(0, TimeUnit.SECONDS).build()
        rf =   Retrofit.Builder()
            .baseUrl("http://100.105.87.39/")
            .addConverterFactory(ScalarsConverterFactory.create())
            .client(client)
            .build()
    }

    @Provides
    fun provideServer(): ServerService {
        return rf.create(ServerService::class.java)
    }
}