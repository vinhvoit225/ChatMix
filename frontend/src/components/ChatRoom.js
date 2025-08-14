import React, { useState, useEffect, useRef, useCallback } from 'react';
import { useAuth } from '../contexts/AuthContext';
import chatService from '../services/chatService';
import ProfileModal from './ProfileModal';
import { Send, Loader2, WifiOff, RefreshCw, Users, Eye } from 'lucide-react';
import toast from 'react-hot-toast';
import CONFIG from '../config/api';

const ChatRoom = ({ onBack }) => {
  const { user, isAuthenticated } = useAuth();
  const [roomCode, setRoomCode] = useState('');
  const [messages, setMessages] = useState([]);
  const [input, setInput] = useState('');
  const [status, setStatus] = useState('connecting'); // connecting, connected, disconnected, error
  const [isLoading, setIsLoading] = useState(true);
  const [profileModal, setProfileModal] = useState({ isOpen: false, username: '' });
  const [onlineUsersCount, setOnlineUsersCount] = useState(0);
  const [peerUser, setPeerUser] = useState(null); // Store peer user profile
  const [queueStatus, setQueueStatus] = useState(null); // Queue status when user is queued
  const messagesEndRef = useRef(null);

  // Fetch online users count
  const fetchOnlineUsersCount = useCallback(async () => {
    try {
      const response = await fetch(`${CONFIG.API_BASE_URL}/users/online`);
      if (response.ok) {
        const onlineUsers = await response.json();
        setOnlineUsersCount(onlineUsers.length);
      }
    } catch (error) {
      console.error('Failed to fetch online users count:', error);
    }
  }, []);

  // Fetch peer user profile
  const fetchPeerProfile = useCallback(async (username) => {
    if (!username || username === user?.username) return;
    
    try {
      const response = await fetch(`${CONFIG.API_BASE_URL}/users/${encodeURIComponent(username)}`);
      if (response.ok) {
        const profile = await response.json();
        setPeerUser(profile);
      }
    } catch (error) {
      console.error('Failed to fetch peer profile:', error);
    }
  }, [user?.username]);

  // Extract username from system message
  const extractUsernameFromSystemMessage = (text) => {
    const joinMatch = text.match(/(.+) đã vào phòng chat/);
    const leaveMatch = text.match(/(.+) đã rời khỏi phòng chat/);
    return joinMatch ? joinMatch[1] : (leaveMatch ? leaveMatch[1] : null);
  };

  // Initialize chat when component mounts
  useEffect(() => {
    if (!isAuthenticated || !user) {
      return;
    }

    let isMounted = true;

    const initializeChat = async () => {
      try {
        setIsLoading(true);
        setStatus('connecting');

        // Step 1: Start chat (get room code or queue status)
        const response = await chatService.startChat(user.username);
        
        if (!isMounted) return;

        if (response.status === 'queued') {
          // User is in queue
          setQueueStatus({
            in_queue: true,
            position: response.position,
            queue_size: response.position || 1
          });
          setStatus('queued');
          setIsLoading(false);
          toast(`Đã thêm vào hàng chờ. Vị trí: ${response.position}`);
          
          // Request notification permission for queue updates
          if ('Notification' in window && Notification.permission === 'default') {
            Notification.requestPermission().then(permission => {
              if (permission === 'granted') {
                toast.success('Sẽ thông báo khi có phòng trống!');
              }
            });
          }
          
          // Start polling queue status
          const queueInterval = setInterval(async () => {
            try {
              const queueCheck = await chatService.startChat(user.username);
              if (queueCheck.status === 'room_assigned') {
                clearInterval(queueInterval);
                // Room assigned, continue with room initialization
                const assignedRoomCode = queueCheck.room;
                setRoomCode(assignedRoomCode);
                setQueueStatus({ in_queue: false, position: 0, queue_size: 0 });
                
                // Show browser notification
                if ('Notification' in window && Notification.permission === 'granted') {
                  new Notification('ChatMix - Phòng chat có sẵn!', {
                    body: `Bạn đã được chỉ định vào phòng ${assignedRoomCode}. Hãy quay lại tab để bắt đầu chat!`,
                    icon: '/favicon.ico',
                    badge: '/favicon.ico',
                    requireInteraction: true,
                    tag: 'room-assigned'
                  });
                }
                
                // Continue with WebSocket connection
                toast.success(`Đã được chỉ định vào room: ${assignedRoomCode}`);
                
                // Initialize WebSocket connection
                try {
                  await chatService.connectToRoom(assignedRoomCode, user.username);
                  setStatus('connected');
                  setIsLoading(false);
                  toast.success(`Đã vào phòng chat thành công!`);
                } catch (error) {
                  console.error('Failed to connect to assigned room:', error);
                  setStatus('error');
                  toast.error('Không thể kết nối tới phòng chat');
                }
              } else if (queueCheck.status === 'queued') {
                setQueueStatus({
                  in_queue: true,
                  position: queueCheck.position,
                  queue_size: queueCheck.position || 1
                });
              }
            } catch (error) {
              console.error('Queue check failed:', error);
            }
          }, 5000); // Check every 5 seconds
          
          return;
        }

        // User got room assignment
        const roomCode = response.room;
        setRoomCode(roomCode);
        console.log('Assigned to room:', roomCode);

        // Step 2: Connect to WebSocket
        await chatService.connectToRoom(roomCode, user.username);
        
        if (!isMounted) return;
        
        setStatus('connected');
        toast.success(`Đã vào phòng ${roomCode}`);

      } catch (error) {
        console.error('Failed to initialize chat:', error);
        if (isMounted) {
          setStatus('error');
          toast.error(error.message || 'Không thể kết nối chat');
        }
      } finally {
        if (isMounted) {
          setIsLoading(false);
        }
      }
    };

    initializeChat();

    return () => {
      isMounted = false;
      chatService.disconnect();
    };
  }, [isAuthenticated, user]);

  // Set up message and status handlers
  useEffect(() => {
    const handleMessage = (message) => {
      setMessages(prev => [...prev, {
        id: `${Date.now()}-${Math.random()}`,
        type: message.type,
        from: message.from,
        text: message.text,
        timestamp: message.timestamp || Date.now(),
        isMe: message.from === user?.username
      }]);

      // Handle system messages to track room users
      if (message.type === 'system') {
        const username = extractUsernameFromSystemMessage(message.text);
        
        if (message.text.includes('đã vào phòng chat')) {
          // Someone joined
          if (username && username !== user?.username) {
            // Fetch their profile
            fetchPeerProfile(username);
          }
        } else if (message.text.includes('đã rời khỏi phòng chat')) {
          // Someone left
          if (username) {
            // Clear peer profile if they were the peer
            if (username !== user?.username && peerUser?.username === username) {
              setPeerUser(null);
            }
          }
        }
        
        // Refresh online users count
        setTimeout(fetchOnlineUsersCount, 1000);
      }

      // Track users from regular messages too
      if (message.type === 'message' && message.from && message.from !== user?.username) {
        // Fetch peer profile if we don't have it yet
        if (!peerUser || peerUser.username !== message.from) {
          fetchPeerProfile(message.from);
        }
      }
    };

    const handleStatusChange = (newStatus) => {
      setStatus(newStatus);
      
      if (newStatus === 'connected') {
        toast.success('Đã kết nối');
      } else if (newStatus === 'disconnected') {
        toast.error('Mất kết nối');
      } else if (newStatus === 'error') {
        toast.error('Lỗi kết nối');
      }
    };

    chatService.onMessage(handleMessage);
    chatService.onStatusChange(handleStatusChange);

    return () => {
      chatService.removeMessageHandler(handleMessage);
      chatService.removeStatusHandler(handleStatusChange);
    };
  }, [user, peerUser, fetchPeerProfile, fetchOnlineUsersCount]);

  // Auto scroll to bottom when new messages arrive (without jumping)
  useEffect(() => {
    if (messagesEndRef.current) {
      // Use scrollTop instead of scrollIntoView to prevent jumping
      const container = messagesEndRef.current.parentElement;
      if (container) {
        container.scrollTop = container.scrollHeight;
      }
    }
  }, [messages]);

  // Fetch online users count periodically
  useEffect(() => {
    // Initial fetch
    fetchOnlineUsersCount();
    
    // Set up interval to refresh every 10 seconds
    const interval = setInterval(fetchOnlineUsersCount, 10000);
    
    return () => clearInterval(interval);
  }, [fetchOnlineUsersCount]);

  // Send message
  const sendMessage = () => {
    if (!input.trim() || !chatService.isConnected()) {
      return;
    }

    try {
      chatService.sendMessage(input);
      setInput('');
    } catch (error) {
      console.error('Failed to send message:', error);
      toast.error('Không thể gửi tin nhắn');
    }
  };

  // Handle key press in input
  const handleKeyPress = (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      sendMessage();
    }
  };

  // Retry connection
  const retry = async () => {
    if (!roomCode || !user) return;

    try {
      setStatus('connecting');
      await chatService.connectToRoom(roomCode, user.username);
      toast.success('Đã kết nối lại');
    } catch (error) {
      console.error('Retry failed:', error);
      toast.error('Không thể kết nối lại');
    }
  };

  // Open profile modal
  const openProfile = (username) => {
    if (username && username !== user?.username) {
      setProfileModal({ isOpen: true, username });
    }
  };

  // Close profile modal
  const closeProfile = () => {
    setProfileModal({ isOpen: false, username: '' });
  };

  // Render loading state
  if (isLoading) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-violet-400 via-purple-500 to-indigo-600 flex items-center justify-center relative overflow-hidden">
        {/* Animated background elements */}
        <div className="absolute inset-0">
          <div className="absolute top-10 left-10 w-72 h-72 bg-pink-300 rounded-full mix-blend-multiply filter blur-xl opacity-70 animate-blob"></div>
          <div className="absolute top-10 right-10 w-72 h-72 bg-yellow-300 rounded-full mix-blend-multiply filter blur-xl opacity-70 animate-blob animation-delay-2000"></div>
          <div className="absolute -bottom-8 left-20 w-72 h-72 bg-blue-300 rounded-full mix-blend-multiply filter blur-xl opacity-70 animate-blob animation-delay-4000"></div>
        </div>
        
        <div className="relative bg-white/90 backdrop-blur-xl rounded-3xl shadow-2xl p-8 text-center border border-white/20">
          <Loader2 className="h-12 w-12 mx-auto text-purple-600 animate-spin mb-4" />
          <h2 className="text-xl font-semibold text-gray-800 mb-2">Đang tìm phòng chat...</h2>
          <p className="text-gray-600">Vui lòng đợi trong giây lát</p>
        </div>
      </div>
    );
  }

  // Render queue state
  if (status === 'queued' && queueStatus?.in_queue) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-violet-400 via-purple-500 to-indigo-600 flex items-center justify-center relative overflow-hidden">
        {/* Animated background elements */}
        <div className="absolute inset-0">
          <div className="absolute top-10 left-10 w-72 h-72 bg-pink-300 rounded-full mix-blend-multiply filter blur-xl opacity-70 animate-blob"></div>
          <div className="absolute top-10 right-10 w-72 h-72 bg-yellow-300 rounded-full mix-blend-multiply filter blur-xl opacity-70 animate-blob animation-delay-2000"></div>
          <div className="absolute -bottom-8 left-20 w-72 h-72 bg-blue-300 rounded-full mix-blend-multiply filter blur-xl opacity-70 animate-blob animation-delay-4000"></div>
        </div>
        
        <div className="relative bg-white/90 backdrop-blur-xl rounded-3xl shadow-2xl p-8 text-center border border-white/20 max-w-md mx-4">
          <div className="w-20 h-20 bg-gradient-to-br from-orange-400 to-pink-500 rounded-full flex items-center justify-center mx-auto mb-6">
            <Users className="h-10 w-10 text-white" />
          </div>
          
          <h2 className="text-2xl font-bold text-gray-800 mb-2">Đang trong hàng chờ</h2>
          <p className="text-gray-600 mb-6">Hiện tại tất cả phòng chat đã đầy. Bạn đang chờ trong hàng đợi.</p>
          
          <div className="bg-gradient-to-r from-orange-50 to-pink-50 rounded-2xl p-4 mb-6">
            <div className="flex items-center justify-center space-x-2 mb-2">
              <div className="w-3 h-3 bg-orange-400 rounded-full animate-pulse"></div>
              <span className="text-lg font-semibold text-orange-700">
                Vị trí #{queueStatus.position}
              </span>
            </div>
            <p className="text-sm text-gray-600">
              Bạn sẽ được vào phòng chat khi có chỗ trống
            </p>
          </div>
          
          <div className="space-y-3">
            <div className="flex items-center justify-between text-sm">
              <span className="text-gray-600">Số người đang chờ:</span>
              <span className="font-medium text-gray-800">{queueStatus.queue_size || queueStatus.position}</span>
            </div>
            <div className="flex items-center justify-between text-sm">
              <span className="text-gray-600">Thời gian chờ ước tính:</span>
              <span className="font-medium text-gray-800">{Math.max(1, Math.ceil(queueStatus.position * 0.5))} phút</span>
            </div>
          </div>
          
          <button 
            onClick={onBack}
            className="mt-6 w-full px-6 py-3 bg-gradient-to-r from-gray-600 to-gray-700 hover:from-gray-700 hover:to-gray-800 text-white font-medium rounded-xl transition-all duration-200"
          >
            Quay lại trang chủ
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-violet-400 via-purple-500 to-indigo-600 p-4 relative overflow-hidden">
      {/* Animated background elements */}
      <div className="absolute inset-0">
        <div className="absolute top-20 left-20 w-96 h-96 bg-pink-300 rounded-full mix-blend-multiply filter blur-xl opacity-30 animate-blob"></div>
        <div className="absolute top-40 right-20 w-96 h-96 bg-yellow-300 rounded-full mix-blend-multiply filter blur-xl opacity-30 animate-blob animation-delay-2000"></div>
        <div className="absolute -bottom-8 left-40 w-96 h-96 bg-blue-300 rounded-full mix-blend-multiply filter blur-xl opacity-30 animate-blob animation-delay-4000"></div>
      </div>
      
      <div className="relative max-w-4xl mx-auto bg-white/95 backdrop-blur-xl rounded-3xl shadow-2xl overflow-hidden border border-white/20">
        
        {/* Header */}
        <div className="px-6 py-4 bg-gradient-to-r from-purple-600 to-blue-600 text-white">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-4">
              <div>
                <div className="text-sm opacity-90">Phòng Chat</div>
                <div className="font-bold text-lg tracking-wide">{roomCode || '...'}</div>
              </div>
              <div className="flex items-center space-x-2">
                <div className={`w-3 h-3 rounded-full ${
                  status === 'connected' ? 'bg-green-400' : 
                  status === 'connecting' ? 'bg-yellow-400 animate-pulse' : 
                  'bg-red-400'
                }`}></div>
                <span className="text-sm">
                  {status === 'connected' ? 'Trực tuyến' : 
                   status === 'connecting' ? 'Đang kết nối...' : 
                   'Mất kết nối'}
                </span>
              </div>
            </div>
            
            <div className="flex items-center space-x-4">
              {/* Online users count in header */}
              <div className="flex items-center space-x-2 bg-white/20 backdrop-blur-sm rounded-xl px-3 py-1">
                <Users className="h-4 w-4" />
                <span className="text-sm font-medium">
                  {onlineUsersCount > 0 ? onlineUsersCount : '...'} online
                </span>
              </div>
              
              <button 
                onClick={onBack}
                className="px-4 py-2 bg-white/20 hover:bg-white/30 rounded-xl text-white font-medium transition-colors"
              >
                Quay lại
              </button>
            </div>
          </div>
        </div>

        {/* Peer User Profile Section */}
        {peerUser && (
          <div className="px-4 sm:px-6 py-3 bg-gradient-to-r from-blue-50 to-purple-50 border-b border-gray-200/50">
            <div className="flex items-center justify-between">
              <div className="flex items-center space-x-3 flex-1 min-w-0">
                <div className="w-10 h-10 bg-gradient-to-br from-purple-500 to-blue-500 rounded-full flex items-center justify-center text-white font-bold text-sm flex-shrink-0">
                  {peerUser.username?.charAt(0)?.toUpperCase() || '👤'}
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center space-x-2">
                    <h3 className="font-semibold text-gray-800 truncate">{peerUser.username}</h3>
                    <div className={`w-2 h-2 rounded-full flex-shrink-0 ${peerUser.is_online ? 'bg-green-400' : 'bg-gray-400'}`}></div>
                  </div>
                  <div className="flex items-center space-x-2 sm:space-x-3 text-xs text-gray-600">
                    {peerUser.age && (
                      <span className="flex-shrink-0">{peerUser.age} tuổi</span>
                    )}
                    {peerUser.gender && peerUser.gender !== 'private' && (
                      <span className="hidden sm:inline flex-shrink-0">
                        {peerUser.gender === 'male' ? '👨 Nam' : 
                         peerUser.gender === 'female' ? '👩 Nữ' : 
                         '🧑 Khác'}
                      </span>
                    )}
                    <span className="text-purple-600 truncate">@{peerUser.username}</span>
                  </div>
                </div>
              </div>
              
              <div className="flex items-center space-x-2 flex-shrink-0">
                {peerUser.bio && (
                  <div className="hidden lg:block max-w-xs text-xs text-gray-600 italic truncate">
                    "{peerUser.bio}"
                  </div>
                )}
                <button
                  onClick={() => openProfile(peerUser.username)}
                  className="flex items-center space-x-1 px-2 sm:px-3 py-1 bg-purple-100 hover:bg-purple-200 text-purple-700 rounded-lg text-xs transition-colors"
                >
                  <Eye className="h-3 w-3" />
                  <span className="hidden sm:inline">Xem thêm</span>
                </button>
              </div>
            </div>
          </div>
        )}

        {/* Messages Area */}
        <div className="h-[70vh] flex flex-col">
          
          {/* Messages Container */}
          <div className="flex-1 overflow-y-auto p-6 space-y-4 bg-gradient-to-b from-gray-50/80 to-white/80 backdrop-blur-sm">
            
            {messages.length === 0 ? (
              <div className="h-full flex flex-col items-center justify-center text-gray-500 space-y-4">
                <div className="text-center">
                  <Users className="h-16 w-16 mb-4 opacity-50 mx-auto" />
                  <h3 className="text-lg font-medium mb-2">
                    {peerUser ? 
                      `Đang chat với ${peerUser.username}!` : 
                      'Chào mừng đến phòng chat!'
                    }
                  </h3>
                  <p className="text-center mb-4">
                    {status === 'connected' ? 
                      (peerUser ? 
                        'Hãy bắt đầu cuộc trò chuyện!' : 
                        'Đang chờ ai đó vào phòng chat...'
                      ) :
                      'Đang chờ kết nối...'
                    }
                  </p>
                </div>
                
                {/* Online users info */}
                {!peerUser && (
                  <div className="bg-white/80 backdrop-blur-sm rounded-2xl p-4 border border-gray-200/50 shadow-lg">
                    <div className="flex items-center justify-center space-x-2">
                      <div className="w-3 h-3 bg-green-400 rounded-full animate-pulse"></div>
                      <span className="text-sm font-medium text-gray-700">
                        {onlineUsersCount > 0 ? (
                          <>
                            <span className="text-green-600 font-bold">{onlineUsersCount}</span> người đang online
                          </>
                        ) : (
                          'Đang tải thông tin...'
                        )}
                      </span>
                    </div>
                    
                    {status === 'connected' && onlineUsersCount > 1 && (
                      <p className="text-xs text-gray-500 text-center mt-2">
                        Có thể sẽ có người vào phòng chat của bạn
                      </p>
                    )}
                  </div>
                )}
              </div>
            ) : (
              <>
                {messages.map((message) => (
                  <div key={message.id} className={`flex ${message.isMe ? 'justify-end' : 'justify-start'}`}>
                    <div className={`max-w-[75%] rounded-2xl px-4 py-3 shadow-lg backdrop-blur-sm ${
                      message.type === 'system' 
                        ? 'bg-gray-200/80 text-gray-700 text-center text-sm mx-auto border border-gray-300/50'
                        : message.isMe 
                        ? 'bg-gradient-to-r from-purple-600 to-blue-600 text-white rounded-br-md border border-purple-300/30'
                        : 'bg-white/90 text-gray-800 border border-gray-200/50 rounded-bl-md'
                    }`}>
                      {message.type !== 'system' && !message.isMe && (
                        <div className="text-xs text-purple-600 mb-1 font-medium">
                          {message.from}
                        </div>
                      )}
                      <div className="whitespace-pre-wrap break-words">{message.text}</div>
                      {message.type !== 'system' && (
                        <div className={`text-xs mt-1 ${message.isMe ? 'text-white/70' : 'text-gray-400'}`}>
                          {new Date(message.timestamp).toLocaleTimeString('vi-VN', { 
                            hour: '2-digit', 
                            minute: '2-digit' 
                          })}
                        </div>
                      )}
                    </div>
                  </div>
                ))}
                <div ref={messagesEndRef} />
              </>
            )}
          </div>

          {/* Connection Status */}
          {status !== 'connected' && (
            <div className="px-6 py-3 bg-yellow-50 border-t border-yellow-200 flex items-center justify-between">
              <div className="flex items-center space-x-2 text-yellow-800">
                <WifiOff className="h-4 w-4" />
                <span className="text-sm">
                  {status === 'connecting' ? 'Đang kết nối...' : 'Mất kết nối'}
                </span>
              </div>
              {status === 'error' && (
                <button 
                  onClick={retry}
                  className="flex items-center space-x-2 px-3 py-1 bg-yellow-200 hover:bg-yellow-300 rounded-lg text-yellow-800 text-sm transition-colors"
                >
                  <RefreshCw className="h-4 w-4" />
                  <span>Thử lại</span>
                </button>
              )}
            </div>
          )}

          {/* Input Area */}
          <div className="p-6 border-t border-gray-200/50 bg-white/90 backdrop-blur-sm">
            <div className="flex items-center space-x-3">
              <input
                type="text"
                value={input}
                onChange={(e) => setInput(e.target.value)}
                onKeyPress={handleKeyPress}
                placeholder="Nhập tin nhắn..."
                disabled={status !== 'connected'}
                className="flex-1 px-4 py-3 rounded-2xl border border-gray-300/50 focus:ring-2 focus:ring-purple-500 focus:border-purple-500 outline-none disabled:bg-gray-100 disabled:cursor-not-allowed bg-white/90 backdrop-blur-sm shadow-sm"
              />
              <button
                onClick={sendMessage}
                disabled={!input.trim() || status !== 'connected'}
                className="flex items-center justify-center w-12 h-12 bg-gradient-to-r from-purple-600 to-blue-600 text-white rounded-2xl hover:from-purple-700 hover:to-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-all shadow-lg hover:shadow-xl transform hover:scale-105"
              >
                <Send className="h-5 w-5" />
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* Profile Modal */}
      <ProfileModal
        username={profileModal.username}
        isOpen={profileModal.isOpen}
        onClose={closeProfile}
      />
    </div>
  );
};

export default ChatRoom;
