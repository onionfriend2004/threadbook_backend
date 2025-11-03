import { IsArray, IsString, ArrayMinSize } from 'class-validator';

export class BatchStatusDto {
  @IsArray()
  @ArrayMinSize(1)
  @IsString({ each: true })
  usernames!: string[];
}
