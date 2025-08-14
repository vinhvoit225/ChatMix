import React, { useState } from 'react';
import { MessageCircle } from 'lucide-react';
import LoginForm from './LoginForm';
import RegisterForm from './RegisterForm';

const AuthPage = ({ onAuthSuccess }) => {
  const [isLoginMode, setIsLoginMode] = useState(true);

  const switchToRegister = () => setIsLoginMode(false);
  const switchToLogin = () => setIsLoginMode(true);

  return (
    <div className="h-screen bg-gradient-to-br from-purple-400 via-pink-500 to-red-500 flex overflow-hidden">
      {/* Centered Auth Forms */}
      <div className="w-full h-full flex items-center justify-center p-4 overflow-y-auto">
        <div className="w-full max-w-md my-auto">
          {/* Logo */}
          <div className="text-center mb-8 text-white">
            <div className="mx-auto h-16 w-16 bg-white bg-opacity-20 backdrop-blur-md rounded-full flex items-center justify-center mb-4">
              <MessageCircle className="h-8 w-8 text-white" />
            </div>
            <h1 className="text-3xl font-bold mb-2">ChatMix</h1>
            <p className="text-white text-opacity-90">Nơi kết nối mọi người</p>
          </div>

          {/* Auth Toggle Tabs */}
          <div className="mb-8">
            <div className="glass-effect rounded-2xl p-2 flex">
              <button
                onClick={switchToLogin}
                className={`flex-1 py-3 px-6 rounded-xl font-semibold transition-all duration-200 ${
                  isLoginMode
                    ? 'bg-white text-gray-800 shadow-lg'
                    : 'text-gray-600 hover:text-gray-800'
                }`}
              >
                Đăng nhập
              </button>
              <button
                onClick={switchToRegister}
                className={`flex-1 py-3 px-6 rounded-xl font-semibold transition-all duration-200 ${
                  !isLoginMode
                    ? 'bg-white text-gray-800 shadow-lg'
                    : 'text-gray-600 hover:text-gray-800'
                }`}
              >
                Đăng ký
              </button>
            </div>
          </div>

          {/* Auth Forms */}
          <div className="relative">
            <div
              className={`transition-all duration-500 ease-in-out ${
                isLoginMode
                  ? 'opacity-100 scale-100 translate-x-0 z-10'
                  : 'opacity-0 scale-95 -translate-x-4 absolute top-0 left-0 w-full pointer-events-none z-0'
              }`}
            >
              <LoginForm 
                onSwitchToRegister={switchToRegister} 
                onAuthSuccess={onAuthSuccess}
              />
            </div>
            
            <div
              className={`transition-all duration-500 ease-in-out ${
                !isLoginMode
                  ? 'opacity-100 scale-100 translate-x-0 z-10'
                  : 'opacity-0 scale-95 translate-x-4 absolute top-0 left-0 w-full pointer-events-none z-0'
              }`}
            >
              <RegisterForm 
                onSwitchToLogin={switchToLogin}
                onAuthSuccess={onAuthSuccess}
              />
            </div>
          </div>

          {/* Footer */}
          <div className="mt-8 text-center text-white text-opacity-70 text-sm">
            <p>
              Bằng việc sử dụng ChatMix, bạn đồng ý với{' '}
              <a href="#" className="text-white hover:underline">Điều khoản dịch vụ</a>
              {' '}và{' '}
              <a href="#" className="text-white hover:underline">Chính sách bảo mật</a>
            </p>
          </div>
        </div>
      </div>

      {/* Background Animation */}
      <div className="fixed inset-0 pointer-events-none overflow-hidden">
        <div className="absolute -top-4 -left-4 w-72 h-72 bg-purple-300 rounded-full mix-blend-multiply filter blur-xl opacity-70 animate-blob"></div>
        <div className="absolute -top-4 -right-4 w-72 h-72 bg-yellow-300 rounded-full mix-blend-multiply filter blur-xl opacity-70 animate-blob animation-delay-2000"></div>
        <div className="absolute -bottom-8 left-20 w-72 h-72 bg-pink-300 rounded-full mix-blend-multiply filter blur-xl opacity-70 animate-blob animation-delay-4000"></div>
      </div>
    </div>
  );
};

export default AuthPage;
