import axios from 'axios';
import CONFIG from '../config/api';

const API_BASE_URL = process.env.REACT_APP_API_URL || CONFIG.API_BASE_URL.replace('/api', '');

// Create axios instance with default config
const apiClient = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add auth token to requests
apiClient.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('access_token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Handle token refresh on 401
apiClient.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;
    
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;
      
      try {
        const refreshToken = localStorage.getItem('refresh_token');
        
        if (refreshToken) {
          const response = await axios.post(`${API_BASE_URL}/api/auth/refresh`, {
            refresh_token: refreshToken
          });
          
          const { access_token } = response.data;
          localStorage.setItem('access_token', access_token);
          
          originalRequest.headers.Authorization = `Bearer ${access_token}`;
          return apiClient(originalRequest);
        }
      } catch (refreshError) {
        // Refresh failed, clear tokens and trigger logout
        localStorage.removeItem('access_token');
        localStorage.removeItem('refresh_token');
        
        // Dispatch custom event to trigger logout in AuthContext
        window.dispatchEvent(new CustomEvent('auth:logout'));
        
        // Show error message
        if (window.toast) {
          window.toast.error('Phiên đăng nhập đã hết hạn. Vui lòng đăng nhập lại.');
        }
        
        return Promise.reject(refreshError);
      }
    }
    
    return Promise.reject(error);
  }
);

class MatchingService {
  // Find match (instant matching)
  async findMatch(preferences = {}) {
    try {
      const response = await apiClient.post('/api/matching/find', {
        preferences: {
          age_range: preferences.ageRange || null,
          gender: preferences.gender || null,
          interests: preferences.interests || [],
          language: preferences.language || 'vi',
          max_wait_time: preferences.maxWaitTime || 30,
        }
      });
      return response.data;
    } catch (error) {
      console.error('Failed to find match:', error);
      throw this.handleError(error);
    }
  }

  // Get matching status
  async getMatchingStatus() {
    try {
      const response = await apiClient.get('/api/matching/status');
      return response.data;
    } catch (error) {
      console.error('Failed to get matching status:', error);
      throw this.handleError(error);
    }
  }

  // Get matching statistics
  async getMatchingStats() {
    try {
      const response = await apiClient.get('/api/matching/stats');
      return response.data;
    } catch (error) {
      console.error('Failed to get matching stats:', error);
      throw this.handleError(error);
    }
  }

  // Join a matched room
  async joinRoom(roomCode) {
    try {
      const response = await apiClient.post(`/api/matching/room/${roomCode}/join`);
      return response.data;
    } catch (error) {
      console.error('Failed to join room:', error);
      throw this.handleError(error);
    }
  }

  // Leave a room
  async leaveRoom(roomCode) {
    try {
      const response = await apiClient.post(`/api/matching/room/${roomCode}/leave`);
      return response.data;
    } catch (error) {
      console.error('Failed to leave room:', error);
      throw this.handleError(error);
    }
  }

  // End a room
  async endRoom(roomCode) {
    try {
      const response = await apiClient.post(`/api/matching/room/${roomCode}/end`);
      return response.data;
    } catch (error) {
      console.error('Failed to end room:', error);
      throw this.handleError(error);
    }
  }

  // Get room statistics
  async getRoomStats(roomCode) {
    try {
      const response = await apiClient.get(`/api/matching/room/${roomCode}/stats`);
      return response.data;
    } catch (error) {
      console.error('Failed to get room stats:', error);
      throw this.handleError(error);
    }
  }

  // Create WebSocket connection for room
  createRoomWebSocket(roomCode, username, onMessage, onError, onClose, onOpen) {
    const wsUrl = `${CONFIG.WS_BASE_URL}/ws/room?username=${encodeURIComponent(username)}&room=${encodeURIComponent(roomCode)}`;
    
    const ws = new WebSocket(wsUrl);
    
    ws.onopen = () => {
      console.log('WebSocket connected to room:', roomCode);
      if (onOpen) onOpen();
    };
    
    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        onMessage(data);
      } catch (error) {
        console.error('Failed to parse WebSocket message:', error);
      }
    };
    
    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
      if (onError) onError(error);
    };
    
    ws.onclose = (event) => {
      console.log('WebSocket connection closed:', event.code, event.reason);
      if (onClose) onClose(event);
    };
    
    return ws;
  }

  // Send message through WebSocket
  sendMessage(ws, content, roomCode) {
    if (ws && ws.readyState === WebSocket.OPEN) {
      const message = {
        type: 'message',
        content: content,
        room_id: roomCode,
        timestamp: new Date().toISOString()
      };
      ws.send(JSON.stringify(message));
      return true;
    }
    return false;
  }

  // Helper method to handle API errors
  handleError(error) {
    if (error.response) {
      // Server responded with error status
      const { status, data } = error.response;
      return {
        status,
        message: data.error || data.message || 'An error occurred',
        details: data
      };
    } else if (error.request) {
      // Request made but no response received
      return {
        status: 0,
        message: 'Network error. Please check your connection.',
        details: error.message
      };
    } else {
      // Something else happened
      return {
        status: -1,
        message: error.message || 'An unexpected error occurred',
        details: error
      };
    }
  }

  // Utility method to format wait time
  formatWaitTime(minutes) {
    if (minutes < 1) return 'Less than 1 minute';
    if (minutes === 1) return '1 minute';
    if (minutes < 60) return `${minutes} minutes`;
    
    const hours = Math.floor(minutes / 60);
    const remainingMinutes = minutes % 60;
    
    if (hours === 1) {
      return remainingMinutes > 0 ? `1 hour ${remainingMinutes} minutes` : '1 hour';
    }
    
    return remainingMinutes > 0 ? `${hours} hours ${remainingMinutes} minutes` : `${hours} hours`;
  }

  // Utility method to validate preferences
  validatePreferences(preferences) {
    const errors = [];
    
    if (preferences.ageRange) {
      const { min, max } = preferences.ageRange;
      if (min < 13 || min > 100) {
        errors.push('Minimum age must be between 13 and 100');
      }
      if (max < 13 || max > 100) {
        errors.push('Maximum age must be between 13 and 100');
      }
      if (min > max) {
        errors.push('Minimum age cannot be greater than maximum age');
      }
    }
    
    if (preferences.interests && preferences.interests.length > 10) {
      errors.push('Maximum 10 interests allowed');
    }
    
    if (preferences.maxWaitTime && (preferences.maxWaitTime < 1 || preferences.maxWaitTime > 60)) {
      errors.push('Max wait time must be between 1 and 60 minutes');
    }
    
    return errors;
  }
}

// Export singleton instance
const matchingService = new MatchingService();
export default matchingService;
