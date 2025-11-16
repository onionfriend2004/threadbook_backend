<<<<<<< Updated upstream
// src/shared/nats/nats.service.ts
=======
>>>>>>> Stashed changes
import {
  JetStreamClient,
  NatsConnection,
  JSONCodec,
  connect,
  Codec,
} from 'nats';
import {
  OnModuleDestroy,
  OnModuleInit,
  Injectable,
  Logger,
} from '@nestjs/common';
<<<<<<< Updated upstream
import { JetStreamClient, NatsConnection, JSONCodec, connect } from 'nats';
=======
>>>>>>> Stashed changes

@Injectable()
export class NatsService implements OnModuleInit, OnModuleDestroy {
  private readonly logger = new Logger(NatsService.name);
  private nc: NatsConnection | null = null;
  private js: JetStreamClient | null = null;
<<<<<<< Updated upstream
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
=======
  private jsonCodec: Codec<unknown> = JSONCodec();

  async onModuleInit() {
    await this.connect();
  }

  async onModuleDestroy() {
    await this.close();
  }

  private async connect() {
    const servers = process.env.NATS_SERVERS?.split(',') || [
      'nats://localhost:4222',
    ];
    const name = process.env.NATS_CLIENT_NAME || 'status-service';

    try {
      this.nc = await connect({
        servers,
        name,
        reconnect: true,
        maxReconnectAttempts: -1,
        reconnectTimeWait: 2000,
      });

      this.js = this.nc.jetstream();

      this.logger.log(`Connected to NATS servers: ${servers.join(', ')}`);

      void this.nc.closed().then((err) => {
        if (err) {
          this.logger.error(
            `NATS connection closed with error: ${err.message}`,
          );
        } else {
          this.logger.log('NATS connection closed');
        }
      });
    } catch (error) {
      this.logger.error('Failed to connect to NATS', error);
      throw error;
    }
  }

  getJetStreamClient(): JetStreamClient {
    if (!this.js) {
      throw new Error('NATS JetStream client is not initialized');
    }
    return this.js;
  }

  getJsonCodec(): Codec<unknown> {
    return this.jsonCodec;
  }

  publish(subject: string, data: unknown): void {
    if (!this.nc) {
      throw new Error('NATS connection is not initialized');
    }
    try {
      const encoded = this.jsonCodec.encode(data);
      void this.nc.publish(subject, encoded);
    } catch (error) {
      this.logger.error(
        `Failed to encode or publish data for NATS subject ${subject}`,
        error,
      );
    }
  }

  async close() {
    if (this.nc && !this.nc.isClosed()) {
      await this.nc.close();
      this.logger.log('NATS connection closed');
>>>>>>> Stashed changes
    }
  }

  async publishJetStream(subject: string, data: unknown): Promise<void> {
    if (!this.js) {
      throw new Error('NATS JetStream client is not initialized');
    }
    const encoded = this.jsonCodec.encode(data);
    await this.js.publish(subject, encoded);
  }

  getConnection(): NatsConnection {
    if (!this.nc) {
      throw new Error('NATS connection is not initialized');
    }
    return this.nc;
  }
}
