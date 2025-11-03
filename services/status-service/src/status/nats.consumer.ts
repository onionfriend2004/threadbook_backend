// src/status/nats.consumer.ts
import { Injectable, Logger, OnModuleInit, OnModuleDestroy } from '@nestjs/common';
import { NatsService } from '../shared/nats/nats.service';
import { StatusService } from './status.service';
import { UserRegisteredEventDto } from './dto/nats-event.dto';

@Injectable()
export class NatsConsumer implements OnModuleInit, OnModuleDestroy {
  private readonly logger = new Logger(NatsConsumer.name);
  private subscription: any = null;

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
      const js = this.natsService.getJetStreamClient();
      
      // Создаем/проверяем stream
      const jsm = await js.jetstreamManager();
      try {
        await jsm.streams.add({
          name: 'user-events-stream',
          subjects: ['user.registered'],
          retention: 'limits',
          max_msgs: 10000,
        });
      } catch (error) {
        // Stream уже существует - это нормально
        if (!error.message?.includes('stream name already in use')) {
          this.logger.warn('Stream creation warning', error);
        }
      }

      // Подписываемся только на user.registered
      this.subscription = await js.subscribe('user.registered', {
        config: {
          durable_name: 'status-service', // Уникальное имя для этого сервиса
          deliver_policy: 'all',
        },
      });

      this.logger.log('NATS consumer started for user.registered events');
      this.processMessages();

    } catch (error) {
      this.logger.error('Failed to start NATS consumer', error);
    }
  }

  private async processMessages() {
    for await (const message of this.subscription) {
      try {
        await this.handleUserRegistered(message);
      } catch (error) {
        this.logger.error('Error processing user.registered event', error);
        message.nak();
      }
    }
  }

  private async handleUserRegistered(message: any) {
    const data = this.natsService.getJsonCodec().decode(message.data);
    
    const event = this.validateEvent(data);
    if (!event) {
      this.logger.warn('Invalid user.registered event data', { data });
      message.term();
      return;
    }

    this.logger.log('Processing user registration', {
      userId: event.user_id,
      username: event.username,
    });

    // Создаем запись статуса для нового пользователя
    await this.statusService.updateCustomStatus(
      event.user_id,
      event.username,
      {
        customStatus: 'Новый пользователь',
        isPrivate: false,
      },
    );

    this.logger.log('User status record created successfully', {
      userId: event.user_id,
      username: event.username,
    });

    message.ack();
  }

  private validateEvent(data: any): UserRegisteredEventDto | null {
    try {
      if (!data || typeof data !== 'object') return null;
      
      // Проверяем обязательные поля для регистрации
      if (typeof data.user_id !== 'number' || data.user_id <= 0) return null;
      if (typeof data.username !== 'string' || !data.username.trim()) return null;
      if (typeof data.email !== 'string' || !data.email.trim()) return null;

      return {
        type: data.type || 0,
        code: data.code || 0,
        email: data.email,
        username: data.username,
        user_id: data.user_id,
      };
    } catch {
      return null;
    }
  }

  private async stopConsumer() {
    if (this.subscription) {
      await this.subscription.destroy();
      this.logger.log('NATS consumer stopped');
    }
  }
}