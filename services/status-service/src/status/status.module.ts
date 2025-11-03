import { StatusController } from './status.controller';
import { StatusService } from './status.service';
import { PrismaModule } from '../shared/prisma/prisma.module';
import { RedisModule } from '../shared/redis/redis.module';
import { Module } from '@nestjs/common';

@Module({
  imports: [PrismaModule, RedisModule],
  controllers: [StatusController],
  providers: [StatusService],
})
export class StatusModule {}
