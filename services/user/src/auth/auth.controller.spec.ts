import { Test, TestingModule } from '@nestjs/testing';
import { UnauthorizedException } from '@nestjs/common';
import { AuthController } from './auth.controller';
import { AuthService } from './auth.service';

describe('AuthController', () => {
  let controller: AuthController;
  let authService: AuthService;

  const mockAuthService = {
    register: jest.fn(),
    login: jest.fn(),
    refreshAccessToken: jest.fn(),
    logout: jest.fn(),
    getProfile: jest.fn(),
  };

  const mockResponse = {
    cookie: jest.fn(),
    clearCookie: jest.fn(),
  } as any;

  beforeEach(async () => {
    const module: TestingModule = await Test.createTestingModule({
      controllers: [AuthController],
      providers: [{ provide: AuthService, useValue: mockAuthService }],
    }).compile();

    controller = module.get<AuthController>(AuthController);
    authService = module.get<AuthService>(AuthService);

    jest.clearAllMocks();
  });

  describe('register', () => {
    const dto = {
      phone: '0612345678',
      username: 'testuser',
      password: 'Password1!',
    };

    it('should register and set cookies', async () => {
      mockAuthService.register.mockResolvedValue({
        access_token: 'jwt-token',
        refresh_token: 'refresh-token',
      });

      const result = await controller.register(dto, mockResponse);

      expect(result).toEqual({ message: 'registered' });
      expect(mockAuthService.register).toHaveBeenCalledWith(dto);
      expect(mockResponse.cookie).toHaveBeenCalledWith(
        'ACCESS_TOKEN',
        'jwt-token',
        expect.objectContaining({ httpOnly: true }),
      );
      expect(mockResponse.cookie).toHaveBeenCalledWith(
        'REFRESH_TOKEN',
        'refresh-token',
        expect.objectContaining({ httpOnly: true }),
      );
    });
  });

  describe('login', () => {
    const dto = { username: 'testuser', password: 'Password1!' };

    it('should login and set cookies', async () => {
      mockAuthService.login.mockResolvedValue({
        access_token: 'jwt-token',
        refresh_token: 'refresh-token',
      });

      const result = await controller.login(dto, mockResponse);

      expect(result).toEqual({ message: 'logged in' });
      expect(mockAuthService.login).toHaveBeenCalledWith(dto);
      expect(mockResponse.cookie).toHaveBeenCalledTimes(2);
    });

    it('should propagate UnauthorizedException from service', async () => {
      mockAuthService.login.mockRejectedValue(
        new UnauthorizedException('Invalid credentials'),
      );

      await expect(controller.login(dto, mockResponse)).rejects.toThrow(
        UnauthorizedException,
      );
    });
  });

  describe('refresh', () => {
    it('should refresh token and set new access cookie', async () => {
      const req = { cookies: { REFRESH_TOKEN: 'valid-refresh-token' } };
      mockAuthService.refreshAccessToken.mockResolvedValue({
        access_token: 'new-jwt-token',
      });

      const result = await controller.refresh(req, mockResponse);

      expect(result).toEqual({ message: 'token refreshed' });
      expect(mockAuthService.refreshAccessToken).toHaveBeenCalledWith(
        'valid-refresh-token',
      );
      expect(mockResponse.cookie).toHaveBeenCalledWith(
        'ACCESS_TOKEN',
        'new-jwt-token',
        expect.objectContaining({ httpOnly: true }),
      );
    });

    it('should throw UnauthorizedException if no refresh token in cookies', async () => {
      const req = { cookies: {} };

      await expect(controller.refresh(req, mockResponse)).rejects.toThrow(
        UnauthorizedException,
      );
      await expect(controller.refresh(req, mockResponse)).rejects.toThrow(
        'No refresh token provided',
      );
    });
  });

  describe('logout', () => {
    it('should logout and clear cookies', async () => {
      const req = { cookies: { REFRESH_TOKEN: 'some-refresh-token' } };
      mockAuthService.logout.mockResolvedValue(undefined);

      const result = await controller.logout(req, mockResponse);

      expect(result).toEqual({ message: 'logged out' });
      expect(mockAuthService.logout).toHaveBeenCalledWith('some-refresh-token');
      expect(mockResponse.clearCookie).toHaveBeenCalledWith('ACCESS_TOKEN');
      expect(mockResponse.clearCookie).toHaveBeenCalledWith('REFRESH_TOKEN');
    });

    it('should logout even without refresh token', async () => {
      const req = { cookies: {} };
      mockAuthService.logout.mockResolvedValue(undefined);

      const result = await controller.logout(req, mockResponse);

      expect(result).toEqual({ message: 'logged out' });
      expect(mockAuthService.logout).toHaveBeenCalledWith(undefined);
    });
  });

  describe('getProfile (me)', () => {
    it('should return user profile', async () => {
      const profile = {
        id: 'user-uuid',
        phone: '0612345678',
        username: 'testuser',
        email: 'test@example.com',
      };
      const req = { user: { userId: 'user-uuid' } };
      mockAuthService.getProfile.mockResolvedValue(profile);

      const result = await controller.getProfile(req);

      expect(result).toEqual(profile);
      expect(mockAuthService.getProfile).toHaveBeenCalledWith('user-uuid');
    });
  });
});
