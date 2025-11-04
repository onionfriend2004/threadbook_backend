import { IsInt, IsString, IsEmail } from 'class-validator';

export class UserRegisteredEventDto {
  @IsInt()
  type!: number;

  @IsInt()
  code!: number;

  @IsEmail()
  email!: string;

  @IsString()
  username!: string;

  @IsInt()
  userId!: number;
}
