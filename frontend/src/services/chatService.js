// Frontend Chat Service - Simple and Clean
import authService from './authService';
import CONFIG from '../config/api';

const WS_BASE_URL = CONFIG.WS_BASE_URL;

class ChatService {
  constructor() {
    this.websocket = null;
    this.currentRoom = null;
    this.username = null;
    this.messageHandlers = [];
    this.statusHandlers = [];
  }

  // Start chat - get room code or queue status
  async startChat(username) {
    try {
      return await authService.apiCall(`/chat/start?username=${encodeURIComponent(username)}`, {
        method: 'POST',
      });
    } catch (error) {
      console.error('Error starting chat:', error);
      throw error;
    }
  }

  // Check queue status
  async checkQueueStatus(username) {
    try {
      return await authService.apiCall(`/chat/queue-status?username=${encodeURIComponent(username)}`);
    } catch (error) {
      console.error('Error checking queue status:', error);
      throw error;
    }
  }

  // Connect to WebSocket for specific room
  connectToRoom(roomCode, username) {
    return new Promise((resolve, reject) => {
      try {
        // Close existing connection
        if (this.websocket) {
          this.websocket.close();
        }

        this.currentRoom = roomCode;
        this.username = username;

        // Get auth token for WebSocket connection
        const token = authService.token;
        if (!token) {
          reject(new Error('Authentication token required for WebSocket connection'));
          return;
        }

        const wsUrl = `${WS_BASE_URL}/ws/chat?room=${encodeURIComponent(roomCode)}&username=${encodeURIComponent(username)}&token=${encodeURIComponent(token)}`;
        this.websocket = new WebSocket(wsUrl);

        this.websocket.onopen = () => {
          console.log('WebSocket connected to room:', roomCode);
          this.notifyStatusChange('connected');
          resolve();
        };

        this.websocket.onmessage = (event) => {
          try {
            const message = JSON.parse(event.data);
            this.handleMessage(message);
          } catch (error) {
            console.error('Error parsing message:', error);
          }
        };

        this.websocket.onclose = (event) => {
          console.log('WebSocket closed:', event.code, event.reason);
          this.notifyStatusChange('disconnected');
        };

        this.websocket.onerror = (error) => {
          console.error('WebSocket error:', error);
          this.notifyStatusChange('error');
          reject(error);
        };

      } catch (error) {
        reject(error);
      }
    });
  }

  // Send message to current room
  sendMessage(text) {
    if (!this.websocket || this.websocket.readyState !== WebSocket.OPEN) {
      throw new Error('WebSocket not connected');
    }

    if (!text.trim()) {
      throw new Error('Message cannot be empty');
    }

    this.websocket.send(text.trim());
  }

  // Disconnect from current room
  disconnect() {
    if (this.websocket) {
      this.websocket.close();
      this.websocket = null;
    }
    this.currentRoom = null;
    this.username = null;
  }

  // Event handlers

  // Add message handler
  onMessage(handler) {
    this.messageHandlers.push(handler);
  }

  // Add status change handler
  onStatusChange(handler) {
    this.statusHandlers.push(handler);
  }

  // Remove message handler
  removeMessageHandler(handler) {
    const index = this.messageHandlers.indexOf(handler);
    if (index > -1) {
      this.messageHandlers.splice(index, 1);
    }
  }

  // Remove status handler
  removeStatusHandler(handler) {
    const index = this.statusHandlers.indexOf(handler);
    if (index > -1) {
      this.statusHandlers.splice(index, 1);
    }
  }

  // Internal methods

  handleMessage(message) {
    // Notify all message handlers
    this.messageHandlers.forEach(handler => {
      try {
        handler(message);
      } catch (error) {
        console.error('Error in message handler:', error);
      }
    });
  }

  notifyStatusChange(status) {
    // Notify all status handlers
    this.statusHandlers.forEach(handler => {
      try {
        handler(status);
      } catch (error) {
        console.error('Error in status handler:', error);
      }
    });
  }

  // Getters
  isConnected() {
    return this.websocket && this.websocket.readyState === WebSocket.OPEN;
  }

  getCurrentRoom() {
    return this.currentRoom;
  }

  getCurrentUsername() {
    return this.username;
  }
}

// Export singleton instance
const chatService = new ChatService();
export default chatService;
