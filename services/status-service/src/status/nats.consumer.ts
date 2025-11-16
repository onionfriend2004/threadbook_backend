<<<<<<< Updated upstream
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
=======
import { UserRegisteredEventDto } from './dto/user-registered.event.dto';
import { plainToInstance } from 'class-transformer';
import { StatusService } from './status.service';
import { NatsService } from '../shared/nats/nats.service';
import { validate } from 'class-validator';
import { 
  JetStreamSubscription,
  RetentionPolicy,
  DeliverPolicy
  StorageType,
  AckPolicy,
  JsMsg, 
>>>>>>> Stashed changes
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
<<<<<<< Updated upstream
  private consumer: any = null; // тип Consumer не экспортируется явно
  private running = false;
=======
  private subscription: JetStreamSubscription | null = null;
  private isRunning = false;
>>>>>>> Stashed changes

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
<<<<<<< Updated upstream
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
=======
      const js = this.natsService.getJetStreamClient();
      const jsm = await js.jetstreamManager();

      // Создаем или обновляем stream
      try {
        await jsm.streams.add({
          name: 'user-events-stream',
          subjects: ['user.registered', 'user.updated.username'],
          retention: RetentionPolicy.Workqueue, // ← Исправлено
          max_msgs: -1,
          max_age: 7 * 24 * 60 * 60 * 1_000_000_000, // 7 дней
          storage: StorageType.File, // ← Исправлено
          num_replicas: 1,
        });
        this.logger.log('Stream user-events-stream created');
      } catch (error: unknown) {
        if (
          error instanceof Error &&
          !error.message.includes('stream name already in use')
        ) {
          this.logger.warn('Stream creation warning', error);
        } else {
          this.logger.log('Stream user-events-stream already exists');
        }
      }

      // Подписываемся на обе темы с wildcard
      this.subscription = await js.subscribe('user.*', {
        config: {
          durable_name: 'status-service-user-events',
          ack_wait: 30 * 1_000_000_000, // 30 сек
          max_deliver: 5,
          ack_policy: AckPolicy.Explicit, // ← Исправлено
          deliver_policy: DeliverPolicy.All, // ← Исправлено
        },
      });

      this.isRunning = true;
      this.logger.log('NATS consumer started for user.* events');
      this.processMessages();
    } catch (error: unknown) {
      this.logger.error(
        'Failed to start NATS consumer',
        error instanceof Error ? error : new Error(String(error)),
      );
      // Оберни setTimeout в void чтобы избежать предупреждения
      void setTimeout(() => this.startConsumer(), 5000);
    }
  }

  private async processMessages() {
    if (!this.subscription) return;

    try {
      for await (const message of this.subscription) {
        if (!this.isRunning) break;

        try {
          await this.handleMessage(message);
        } catch (error: unknown) {
          this.logger.error('Unexpected error in message handler', {
            error: error instanceof Error ? error.message : String(error),
            sequence: message.seq,
          });
          message.nak(10_000_000_000); // 10 сек задержки
        }
      }
    } catch (error: unknown) {
      if (this.isRunning) {
        this.logger.error(
          'Message processing loop failed',
          error instanceof Error ? error : new Error(String(error)),
        );
        void setTimeout(() => this.startConsumer(), 5000);
>>>>>>> Stashed changes
      }
    }
  }

<<<<<<< Updated upstream
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
=======
  private async handleMessage(message: JsMsg) {
    switch (message.subject) {
      case 'user.registered':
        await this.handleUserRegistered(message);
        break;
      case 'user.updated.username':
        await this.handleUsernameUpdated(message);
        break;
      default:
        this.logger.warn('Unknown subject, terminating message', {
          subject: message.subject,
        });
        message.term();
    }
  }

  private async handleUserRegistered(message: JsMsg) {
    let unknown;
    try {
      // ← Объявили переменную data!
      let data = this.natsService.getJsonCodec().decode(message.data);
      this.logger.debug('Received user.registered event', data);
    } catch (error: unknown) {
      this.logger.warn('Failed to decode message as JSON', {
        error: error instanceof Error ? error.message : String(error),
        data: message.data?.toString(),
      });
      message.term();
      return;
    }

    const transformedData = {
      type: data.type,
      verifyCode: data.verify_code,
      email: data.email_to,
      username: data.username,
      userId: data.user_id,
    };

    const eventDto = plainToInstance(UserRegisteredEventDto, transformedData);
    const errors = await validate(eventDto);

    if (errors.length > 0) {
      this.logger.warn('Validation failed for user.registered event', {
        errors: errors.flatMap((err) => Object.values(err.constraints || {})),
        receivedData: data,
      });
      message.term();
      return;
    }

    try {
      await this.statusService.createUserStatusRecord({
        userId: eventDto.userId,
        username: eventDto.username,
        email: eventDto.email,
      });

      this.logger.log('User status record created successfully', {
        userId: eventDto.userId,
        username: eventDto.username,
      });
      message.ack();
    } catch (error: unknown) {
      this.logger.error('Failed to create user status record', {
        userId: eventDto.userId,
        error: error instanceof Error ? error.message : String(error),
      });
      message.nak(5_000_000_000);
    }
  }

  private async handleUsernameUpdated(message: JsMsg) {
    let  unknown;
    try {
      // ← Объявили переменную data!
       = this.natsService.getJsonCodec().decode(message.data);
    } catch (error: unknown) {
      this.logger.warn('Failed to decode username update as JSON', {
        error: error instanceof Error ? error.message : String(error),
      });
      message.term();
      return;
    }

    const userId = data.user_id;
    const newUsername = data.new_username;

    if (
      typeof userId !== 'number' ||
      typeof newUsername !== 'string' ||
      newUsername.length === 0
    ) {
      this.logger.warn('Invalid username update data', { data });
      message.term();
      return;
    }

    try {
      await this.statusService.updateUsername(userId, newUsername);
      this.logger.log('Username updated successfully', { userId, newUsername });
      message.ack();
    } catch (error: unknown) {
      this.logger.error('Failed to update username', {
        userId,
        error: error instanceof Error ? error.message : String(error),
      });
      message.nak(5_000_000_000);
>>>>>>> Stashed changes
    }
  }

  private async stopConsumer() {
<<<<<<< Updated upstream
    this.running = false;
    if (this.consumer) {
      await this.consumer.delete();
      this.logger.log('NATS consumer stopped');
    }
  }
}
=======
    this.isRunning = false;
    if (this.subscription) {
      try {
        await this.subscription.drain();
        await this.subscription.destroy();
        this.logger.log('NATS consumer stopped gracefully');
      } catch (error: unknown) {
        this.logger.warn(
          'Error stopping NATS consumer',
          error instanceof Error ? error : new Error(String(error)),
        );
      }
      this.subscription = null;
    }
  }

  isHealthy(): boolean {
    return this.isRunning && this.subscription !== null;
  }
}
>>>>>>> Stashed changes
