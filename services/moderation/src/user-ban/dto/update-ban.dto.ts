import { IsDateString, IsOptional, IsString, MaxLength } from 'class-validator';

export class UpdateBanDto {
  @IsOptional()
  @IsString({ message: 'La raison doit être une chaîne de caractères' })
  @MaxLength(500, { message: 'La raison ne doit pas dépasser 500 caractères' })
  reason?: string;

  @IsOptional()
  @IsDateString(
    {},
    { message: "La date d'expiration doit être une date valide (ISO 8601)" },
  )
  expiresAt?: string;
}
