import React, { useState, useEffect } from 'react';
import { MessageCircle } from 'lucide-react';
import { Toaster } from 'react-hot-toast';
import toast from 'react-hot-toast';
import { AuthProvider, useAuth } from './contexts/AuthContext';
import AuthPage from './components/auth/AuthPage';
import HomePage from './components/HomePage';
import ChatRoom from './components/ChatRoom';
// Matching/Room components removed with WS cleanup
import './App.css';

// Main App Component
function AppContent() {
  const { isAuthenticated, isLoading } = useAuth();
  const [currentPage, setCurrentPage] = useState('home'); // home, auth, chat

  // Expose toast to window for use in API interceptors
  useEffect(() => {
    window.toast = toast;
    return () => {
      delete window.toast;
    };
  }, []);

  // After login, go to chat if user initiated chat; otherwise stay on home
  useEffect(() => {
    if (isAuthenticated && currentPage === 'auth') {
      setCurrentPage('chat');
    }
  }, [isAuthenticated, currentPage]);

  // Navigation handlers
  const handleStartChat = () => {
    if (isAuthenticated) setCurrentPage('chat');
    else setCurrentPage('auth');
  };
  const handleAuthSuccess = () => setCurrentPage('chat');

  // Show loading screen while checking authentication
  if (isLoading) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-purple-400 via-pink-500 to-red-500 flex items-center justify-center">
        <div className="glass-effect rounded-3xl p-8 text-center animate-bounce-in">
          <MessageCircle className="mx-auto h-16 w-16 text-primary-600 mb-4 animate-pulse" />
          <h1 className="text-3xl font-bold text-gray-800 mb-2">ChatMix</h1>
          <p className="text-gray-600">Đang tải...</p>
        </div>
      </div>
    );
  }

  // Route rendering
  if (!isAuthenticated) {
    switch (currentPage) {
      case 'auth':
        return <AuthPage onAuthSuccess={handleAuthSuccess} />;
      default:
        return <HomePage onGetStarted={handleStartChat} />;
    }
  }

  // Authenticated routes
  switch (currentPage) {
    case 'chat':
      return <ChatRoom onBack={() => setCurrentPage('home')} />;
    default:
      return <HomePage onGetStarted={handleStartChat} />;
  }
}

// App wrapper with AuthProvider
function App() {
  return (
    <AuthProvider>
      <AppContent />
      <Toaster 
        position="top-right"
        toastOptions={{
          duration: 4000,
          style: {
            background: '#363636',
            color: '#fff',
            borderRadius: '12px',
          },
        }}
      />
    </AuthProvider>
  );
}

export default App;
