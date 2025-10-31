import { Module } from '@nestjs/common';
import { StatusModule } from './status/status.module';
import { PostgresModule } from './shared/postgres/postgres.module';
import { RedisModule } from './shared/redis/redis.module';

@Module({
  imports: [StatusModule, PostgresModule, RedisModule],
  controllers: [],
  providers: [],
})
export class AppModule {}
