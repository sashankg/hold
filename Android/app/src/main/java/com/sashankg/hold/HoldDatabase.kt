package com.sashankg.hold

import android.content.Context
import androidx.room.Database
import androidx.room.Room
import androidx.room.RoomDatabase
import androidx.room.TypeConverter
import androidx.room.TypeConverters
import androidx.work.ListenableWorker
import com.sashankg.hold.model.Media
import com.sashankg.hold.model.MediaDao
import com.sashankg.hold.model.MediaModule
import dagger.Module
import dagger.Provides
import dagger.hilt.InstallIn
import dagger.hilt.android.qualifiers.ApplicationContext
import dagger.hilt.components.SingletonComponent
import java.util.Date
import javax.inject.Inject

@Database(entities = [Media::class], version = 1, exportSchema = false)
@TypeConverters(Converters::class)
abstract class HoldDatabase: RoomDatabase() {
    abstract fun mediaDao(): MediaDao
}

@Module
@InstallIn(SingletonComponent::class)
object DbModule {
    @Provides
    fun provideDb(
        @ApplicationContext context: Context
    ): HoldDatabase {
        return Room.databaseBuilder(
            context,
            HoldDatabase::class.java, "hold-database"
        ).build()
    }
}

class Converters {
    @TypeConverter
    fun fromTimestamp(value: Long?): Date? {
        return value?.let { Date(it) }
    }

    @TypeConverter
    fun dateToTimestamp(date: Date?): Long? {
        return date?.time?.toLong()
    }
}