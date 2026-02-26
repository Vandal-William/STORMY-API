import {
  IsEnum,
  IsOptional,
  IsString,
  IsUUID,
  MaxLength,
} from 'class-validator';
import { ReportReason } from '@prisma/client';

export class CreateReportDto {
  @IsUUID('4', { message: "L'identifiant du reporter doit être un UUID valide" })
  reporterId: string;

  @IsOptional()
  @IsUUID('4', { message: "L'identifiant de l'utilisateur signalé doit être un UUID valide" })
  reportedUserId?: string;

  @IsOptional()
  @IsUUID('4', { message: "L'identifiant du message signalé doit être un UUID valide" })
  reportedMessageId?: string;

  @IsOptional()
  @IsUUID('4', { message: "L'identifiant de la conversation doit être un UUID valide" })
  conversationId?: string;

  @IsEnum(ReportReason, {
    message: `La raison doit être l'une des suivantes : ${Object.values(ReportReason).join(', ')}`,
  })
  reason: ReportReason;

  @IsOptional()
  @IsString({ message: 'La description doit être une chaîne de caractères' })
  @MaxLength(1000, { message: 'La description ne doit pas dépasser 1000 caractères' })
  description?: string;
}
