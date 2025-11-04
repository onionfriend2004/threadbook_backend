import { StatusService } from './status.service';
import { NatsService } from '../shared/nats/nats.service';
import {
  JetStreamManager,
  JetStreamClient,
  RetentionPolicy,
  DeliverPolicy,
  consumerOpts,
  AckPolicy,
  JsMsg,
} from 'nats';
import {
  OnModuleDestroy,
  OnModuleInit,
  Injectable,
  Logger,
} from '@nestjs/common';

@Injectable()
export class NatsConsumer implements OnModuleInit, OnModuleDestroy {
  private readonly logger = new Logger(NatsConsumer.name);
  private consumer: any = null; // тип Consumer не экспортируется явно
  private running = false;

  constructor(
    private readonly natsService: NatsService,
    private readonly statusService: StatusService,
  ) {}

  async onModuleInit() {
    await this.startConsumer();
  }

  async onModuleDestroy() {
    await this.stopConsumer();
  }

  private async startConsumer() {
    try {
      const js: JetStreamClient = this.natsService.getJetStream();
      const jsm: JetStreamManager = await js.jetstreamManager();

      try {
        await jsm.streams.add({
          name: 'user_events',
          subjects: ['user.registered'],
          retention: RetentionPolicy.Workqueue,
          max_age: 7 * 24 * 3600 * 1_000_000_000,
        });
      } catch (err: any) {
        if (!err.message?.includes('already exists')) {
          this.logger.warn('Stream issue', err);
        }
      }

      this.consumer = await jsm.consumers.add('user_events', {
        durable_name: 'status-service-consumer',
        ack_policy: AckPolicy.Explicit,
        deliver_policy: DeliverPolicy.New,
      });
      this.running = true;
      this.logger.log('Started NATS JetStream consumer');
      this.consumeMessages();
    } catch (error) {
      this.logger.error('Failed to start NATS consumer', error);
    }
  }

  private async consumeMessages() {
    if (!this.consumer || !this.running) return;

    try {
      const msgs: JsMsg[] = await this.consumer.pull({
        batch: 1,
        expires: 5000,
      });
      for (const msg of msgs) {
        await this.processMessage(msg);
      }
    } catch (err) {
      if (this.running) {
        this.logger.debug('Pull timeout or no messages');
      }
    } finally {
      if (this.running) {
        setImmediate(() => this.consumeMessages());
      }
    }
  }

  private async processMessage(msg: JsMsg) {
    try {
      const data = this.natsService.getJsonCodec().decode(msg.data);

      if (
        !data ||
        typeof (data as any).user_id !== 'number' ||
        typeof (data as any).username !== 'string' ||
        (data as any).user_id <= 0 ||
        !(data as any).username.trim()
      ) {
        this.logger.warn('Invalid user.registered event', { data });
        msg.term();
        return;
      }

      const userId = (data as any).user_id;
      const username = (data as any).username;

      await this.statusService.updateCustomStatus(userId, username, {
        customStatus: 'Новый пользователь',
        isPrivate: false,
      });

      msg.ack();
    } catch (error) {
      this.logger.error('Error processing message', error);
      msg.nak();
    }
  }

  private async stopConsumer() {
    this.running = false;
    if (this.consumer) {
      await this.consumer.delete();
      this.logger.log('NATS consumer stopped');
    }
  }
}
