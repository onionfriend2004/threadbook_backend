import 'fastify';

declare module 'fastify' {
  interface FastifyRequest {
    user?: {
      userId: number;
      username: string;
    };
  }
}
