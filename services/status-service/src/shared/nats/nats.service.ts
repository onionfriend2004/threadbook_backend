// src/shared/nats/nats.service.ts
import {
  Injectable,
  Logger,
  OnModuleInit,
  OnModuleDestroy,
} from '@nestjs/common';
import { JetStreamClient, NatsConnection, JSONCodec, connect } from 'nats';

@Injectable()
export class NatsService implements OnModuleInit, OnModuleDestroy {
  private readonly logger = new Logger(NatsService.name);
  private nc: NatsConnection | null = null;
  private js: JetStreamClient | null = null;
  private readonly jsonCodec = JSONCodec();

  async onModuleInit() {
    await this.connect();
  }

  async connect(): Promise<void> {
    const servers = (process.env.NATS_SERVERS || 'localhost:4222').split(',');
    this.nc = await connect({ servers, name: 'status-service' });
    this.js = this.nc.jetstream();
    this.logger.log('Connected to NATS JetStream');
  }

  getJetStream(): JetStreamClient {
    if (!this.js) throw new Error('NATS JetStream not connected');
    return this.js;
  }

  getJsonCodec() {
    return this.jsonCodec;
  }

  async onModuleDestroy() {
    if (this.nc) {
      await this.nc.close();
    }
  }
}
