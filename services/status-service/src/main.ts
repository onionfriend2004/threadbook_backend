import { ValidationPipe } from '@nestjs/common';
import { fastifyCookie } from '@fastify/cookie';
import { NestFactory } from '@nestjs/core';
import { AppModule } from './app.module';
import {
  NestFastifyApplication,
  FastifyAdapter,
} from '@nestjs/platform-fastify';

async function bootstrap() {
  const app = await NestFactory.create<NestFastifyApplication>(
    AppModule,
    new FastifyAdapter({ logger: true }),
  );

  // JSON
  app.useBodyParser('json');

  // Куки
  await app.register(fastifyCookie, {
    secret: process.env.COOKIE_SECRET || 'fallback-secret-please-change',
  });

  app.useGlobalPipes(
    new ValidationPipe({ whitelist: true, forbidNonWhitelisted: true }),
  );

  await app.listen(3000, '0.0.0.0');
}

void (async () => {
  try {
    await bootstrap();
  } catch (error) {
    console.error('Failed to start application:', error);
    process.exit(1);
  }
})();
