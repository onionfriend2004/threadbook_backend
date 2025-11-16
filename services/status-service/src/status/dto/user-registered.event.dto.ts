<<<<<<< Updated upstream
import { IsInt, IsString, IsEmail } from 'class-validator';
=======
import { IsEmail, IsInt, IsNumber, IsString, MaxLength } from 'class-validator';
>>>>>>> Stashed changes

export class UserRegisteredEventDto {
  @IsInt()
  type!: number;

  @IsInt()
<<<<<<< Updated upstream
  code!: number;

  @IsEmail()
  email!: string;

  @IsString()
=======
  verifyCode!: number; // verify_code

  @IsEmail()
  email!: string; // email_to

  @IsString()
  @MaxLength(32)
>>>>>>> Stashed changes
  username!: string;

  @IsNumber()
  @IsInt()
<<<<<<< Updated upstream
  userId!: number;
=======
  userId!: number; // user_id
>>>>>>> Stashed changes
}
