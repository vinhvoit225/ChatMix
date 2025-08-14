// Authentication Service
import CONFIG from '../config/api';

const API_BASE_URL = CONFIG.API_BASE_URL;

class AuthService {
  constructor() {
    // Migrate old token keys to new keys for consistency
    this.migrateTokenKeys();
    
    this.token = localStorage.getItem('access_token');
    this.refreshToken = localStorage.getItem('refresh_token');
  }

  // Migrate old localStorage keys to new keys
  migrateTokenKeys() {
    const oldAccessToken = localStorage.getItem('authToken');
    const oldRefreshToken = localStorage.getItem('refreshToken');
    
    if (oldAccessToken && !localStorage.getItem('access_token')) {
      localStorage.setItem('access_token', oldAccessToken);
      localStorage.removeItem('authToken');
      console.log('🔄 [Auth] Migrated access token to new key');
    }
    
    if (oldRefreshToken && !localStorage.getItem('refresh_token')) {
      localStorage.setItem('refresh_token', oldRefreshToken);
      localStorage.removeItem('refreshToken');
      console.log('🔄 [Auth] Migrated refresh token to new key');
    }
  }

  // Helper method to make API calls
  async apiCall(endpoint, options = {}) {
    const url = `${API_BASE_URL}${endpoint}`;
    const config = {
      headers: {
        'Content-Type': 'application/json',
        ...options.headers,
      },
      ...options,
    };

    // Add authorization header if token exists
    if (this.token && !options.skipAuth) {
      config.headers.Authorization = `Bearer ${this.token}`;
    }

    try {
      const response = await fetch(url, config);
      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || `HTTP error! status: ${response.status}`);
      }

      return data;
    } catch (error) {
      console.error(`API call failed for ${endpoint}:`, error);
      throw error;
    }
  }

  // Generate captcha
  async generateCaptcha() {
    return this.apiCall('/auth/captcha', { skipAuth: true });
  }

  // Register new user
  async register(userData) {
    const response = await this.apiCall('/auth/register', {
      method: 'POST',
      body: JSON.stringify(userData),
      skipAuth: true,
    });

    if (response.token) {
      this.setTokens(response.token, response.refresh_token);
    }

    return response;
  }

  // Login user
  async login(credentials) {
    const response = await this.apiCall('/auth/login', {
      method: 'POST',
      body: JSON.stringify(credentials),
      skipAuth: true,
    });

    if (response.token) {
      this.setTokens(response.token, response.refresh_token);
    }

    return response;
  }

  // Logout user
  async logout() {
    try {
      await this.apiCall('/auth/logout', {
        method: 'POST',
      });
    } catch (error) {
      console.error('Logout API call failed:', error);
    } finally {
      this.clearTokens();
    }
  }

  // Refresh token
  async refreshAccessToken() {
    if (!this.refreshToken) {
      throw new Error('No refresh token available');
    }

    try {
      const response = await this.apiCall('/auth/refresh', {
        method: 'POST',
        body: JSON.stringify({ refresh_token: this.refreshToken }),
        skipAuth: true,
      });

      this.setTokens(response.token, response.refresh_token);
      return response;
    } catch (error) {
      this.clearTokens();
      throw error;
    }
  }

  // Get user profile
  async getProfile() {
    return this.apiCall('/auth/profile');
  }

  // Update user profile
  async updateProfile(profileData) {
    return this.apiCall('/auth/profile', {
      method: 'PUT',
      body: JSON.stringify(profileData),
    });
  }

  // Change password
  async changePassword(passwordData) {
    return this.apiCall('/auth/change-password', {
      method: 'POST',
      body: JSON.stringify(passwordData),
    });
  }

  // Revoke all sessions
  async revokeAllSessions() {
    return this.apiCall('/auth/revoke-sessions', {
      method: 'POST',
    });
  }

  // Token management
  setTokens(accessToken, refreshToken) {
    this.token = accessToken;
    this.refreshToken = refreshToken;
    localStorage.setItem('access_token', accessToken);
    localStorage.setItem('refresh_token', refreshToken);
  }

  clearTokens() {
    this.token = null;
    this.refreshToken = null;
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
  }

  // Check if user is authenticated
  isAuthenticated() {
    return !!this.token;
  }

  // Get current user from token (basic client-side parsing)
  getCurrentUser() {
    if (!this.token) return null;

    try {
      const payload = JSON.parse(atob(this.token.split('.')[1]));
      return {
        id: payload.user_id,
        username: payload.username,
        email: payload.email,
        exp: payload.exp,
      };
    } catch (error) {
      console.error('Failed to parse token:', error);
      return null;
    }
  }

  // Check if token is expired
  isTokenExpired() {
    const user = this.getCurrentUser();
    if (!user) return true;

    return Date.now() >= user.exp * 1000;
  }

  // Auto-refresh token if needed
  async ensureValidToken() {
    if (!this.token) {
      throw new Error('No token available');
    }

    if (this.isTokenExpired()) {
      await this.refreshAccessToken();
    }
  }
}

// Create singleton instance
const authService = new AuthService();
export default authService;
