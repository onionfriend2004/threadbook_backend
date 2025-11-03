import { StatusModule } from './status/status.module';
import { PrismaModule } from './shared/prisma/prisma.module';
import { RedisModule } from './shared/redis/redis.module';
import { AuthModule } from './auth/auth.module';
import { NatsModule } from './shared/nats/nats.module';
import { Module } from '@nestjs/common';

@Module({
  imports: [StatusModule, PrismaModule, RedisModule, AuthModule, NatsModule],
})
export class AppModule {}
