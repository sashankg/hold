package com.sashankg.hold

import dagger.Module
import dagger.Provides
import dagger.hilt.InstallIn
import dagger.hilt.components.SingletonComponent
import okhttp3.MultipartBody
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

interface ServerService {
    @Multipart
    @Headers("Connection: keep-alive")
    @POST("/upload")
    fun upload(@Part("file") file: RequestBody): Call<String>
}

@Module
@InstallIn(SingletonComponent::class)
object ServerModule {
    val rf: Retrofit = Retrofit.Builder()
        .baseUrl("http://100.99.93.94:3000/")
        .addConverterFactory(ScalarsConverterFactory.create())
        .build()

    @Provides
    fun provideServer(): ServerService {
        return rf.create(ServerService::class.java)
    }
}