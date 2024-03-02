package com.sashankg.hold.model

import androidx.room.Dao
import androidx.room.Entity
import androidx.room.Insert
import androidx.room.PrimaryKey
import androidx.room.TypeConverter
import androidx.work.ListenableWorker
import com.sashankg.hold.BackupWorker
import com.sashankg.hold.HoldDatabase
import com.sashankg.hold.WatcherWorker
import dagger.Binds
import dagger.Module
import dagger.Provides
import dagger.hilt.InstallIn
import dagger.hilt.components.SingletonComponent
import java.util.Date


@Entity
data class Media(
    @PrimaryKey val id: Long,
    val foundAt: Date?,
    val uploadedAt: Date?,
)

@Dao
interface MediaDao {
    @Insert
    suspend fun insertAll(media: List<Media>)

    @Insert
    suspend fun insert(media: Media)
}

@Module
@InstallIn(SingletonComponent::class)
object MediaModule {
    @Provides
    fun provideMediaDao(
        db: HoldDatabase
    ): MediaDao {
        return db.mediaDao()
    }
}


