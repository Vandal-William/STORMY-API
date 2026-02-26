import {
  IsDateString,
  IsOptional,
  IsString,
  IsUUID,
  MaxLength,
} from 'class-validator';

export class CreateBanDto {
  @IsUUID('4', { message: "L'identifiant de l'utilisateur doit être un UUID valide" })
  userId: string;

  @IsOptional()
  @IsString({ message: 'La raison doit être une chaîne de caractères' })
  @MaxLength(500, { message: 'La raison ne doit pas dépasser 500 caractères' })
  reason?: string;

  @IsOptional()
  @IsDateString({}, { message: "La date d'expiration doit être une date valide (ISO 8601)" })
  expiresAt?: string;
}
