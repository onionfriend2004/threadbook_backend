import { IsOptional, IsString, MaxLength } from 'class-validator';

export class UpdateStatusDto {
  @IsOptional()
  @IsString()
  @MaxLength(128)
  customStatus?: string;

  @IsOptional()
  isPrivate?: boolean;
}
