import { StatusController } from './status.controller';
import { StatusService } from './status.service';
import { NatsConsumer } from './nats.consumer';
import { PrismaModule } from '../shared/prisma/prisma.module';
import { RedisModule } from '../shared/redis/redis.module';
import { NatsModule } from '../shared/nats/nats.module';
import { Module } from '@nestjs/common';

@Module({
  imports: [PrismaModule, RedisModule, NatsModule],
  providers: [StatusService, NatsConsumer],
  controllers: [StatusController],
  exports: [StatusService],
})
export class StatusModule {}
