import { RedisModule } from '../shared/redis/redis.module';
import { AuthService } from './auth.service';
import { AuthGuard } from './auth.guard';
import { Module } from '@nestjs/common';

@Module({
  imports: [RedisModule],
  providers: [AuthService, AuthGuard],
  exports: [AuthService, AuthGuard],
})
export class AuthModule {}
