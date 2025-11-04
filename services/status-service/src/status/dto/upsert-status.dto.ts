import { IsOptional, IsString, IsBoolean } from 'class-validator';

export class UpsertStatusDto {
  @IsOptional()
  @IsString()
  customStatus?: string;

  @IsOptional()
  @IsBoolean()
  isPrivate?: boolean;
}
