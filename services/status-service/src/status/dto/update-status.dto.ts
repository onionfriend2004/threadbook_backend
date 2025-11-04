import { IsOptional, IsString, IsBoolean, MaxLength } from 'class-validator';

export class UpdateStatusDto {
  @IsOptional()
  @IsString()
  @MaxLength(128)
  customStatus?: string;

  @IsOptional()
  @IsBoolean()
  isPrivate?: boolean;
}
