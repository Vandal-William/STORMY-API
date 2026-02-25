import {
  IsEmail,
  IsOptional,
  IsString,
  MaxLength,
  MinLength,
} from 'class-validator';

export class UpdateProfileDto {
  @IsOptional()
  @IsString({
    message: "Le nom d'utilisateur doit être une chaîne de caractères",
  })
  @MinLength(3, {
    message: "Le nom d'utilisateur doit contenir au moins 3 caractères",
  })
  @MaxLength(50, {
    message: "Le nom d'utilisateur ne doit pas dépasser 50 caractères",
  })
  username?: string;

  @IsOptional()
  @IsEmail({}, { message: "L'adresse email n'est pas valide" })
  @MaxLength(255, {
    message: "L'email ne doit pas dépasser 255 caractères",
  })
  email?: string;

  @IsOptional()
  @IsString({ message: "L'URL de l'avatar doit être une chaîne de caractères" })
  @MaxLength(500, {
    message: "L'URL de l'avatar ne doit pas dépasser 500 caractères",
  })
  avatarUrl?: string;

  @IsOptional()
  @IsString({
    message: 'Le champ "à propos" doit être une chaîne de caractères',
  })
  @MaxLength(500, {
    message: 'Le champ "à propos" ne doit pas dépasser 500 caractères',
  })
  about?: string;
}
