import {
  IsEmail,
  IsOptional,
  IsString,
  Matches,
  MinLength,
} from 'class-validator';

export class RegisterDto {
  @IsString()
  @MinLength(6)
  phone: string;

  @IsString()
  @MinLength(3)
  username: string;

  @IsString()
  @MinLength(8, {
    message: 'Le mot de passe doit contenir au moins 8 caractères',
  })
  @Matches(/(?=.*\d)(?=.*[!@#$%^&*()_+\-=[\]{};':"\\|,.<>/?])/, {
    message:
      'Le mot de passe doit contenir au moins un chiffre et un caractère spécial',
  })
  password: string;

  @IsOptional()
  @IsEmail()
  email?: string;
}
