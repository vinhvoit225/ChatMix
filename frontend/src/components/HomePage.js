import React, { useState } from 'react';
import { 
  MessageCircle, 
  Users, 
  Shield, 
  Sparkles, 
  Globe, 
  Heart,
  ArrowRight,
  UserX,
  EyeOff,
  X,
  LogOut,
  ChevronDown,
  Pencil
} from 'lucide-react';
import { useAuth } from '../contexts/AuthContext';
import EditProfileModal from './EditProfileModal';

const HomePage = ({ onGetStarted }) => {
  const { user, logout } = useAuth();
  const [menuOpen, setMenuOpen] = useState(false);
  const [showEditProfile, setShowEditProfile] = useState(false);

  const displayName = user?.username || 'User';
  const initial = (displayName || 'U').charAt(0).toUpperCase();

  const handleLogout = async () => {
    try {
      await logout();
    } finally {
      setMenuOpen(false);
    }
  };

  const features = [
    {
      icon: <MessageCircle className="h-8 w-8" />,
      title: "Chat Ẩn Danh",
      description: "Thoải mái là chính mình"
    },
    {
      icon: <Users className="h-8 w-8" />,
      title: "Cộng đồng thân thiện",
      description: "Kết nối với những người có cùng sở thích và đam mê"
    },
    {
      icon: <Shield className="h-8 w-8" />,
      title: "An toàn & Bảo mật",
      description: "Hệ thống bảo mật tiên tiến và kiểm duyệt thông minh"
    },
    {
      icon: <Sparkles className="h-8 w-8" />,
      title: "Giao diện hiện đại",
      description: "Thiết kế đẹp mắt, dễ sử dụng và tối ưu cho mọi thiết bị"
    }
  ];

  return (
    <div className="min-h-screen bg-gradient-to-br from-purple-50 via-white to-blue-50">
      {/* Header */}
      <header className="bg-white/80 backdrop-blur-md border-b border-gray-200/50 sticky top-0 z-50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center py-4">
            {/* Logo */}
            <div className="flex items-center space-x-2">
              <div className="h-10 w-10 bg-gradient-to-r from-purple-600 to-blue-600 rounded-full flex items-center justify-center">
                <MessageCircle className="h-6 w-6 text-white" />
              </div>
              <span className="text-2xl font-bold bg-gradient-to-r from-purple-600 to-blue-600 bg-clip-text text-transparent">
                ChatMix
              </span>
            </div>
            {/* Right side: Profile (if logged in) or CTA */}
            {user ? (
              <div className="relative">
                <button
                  onClick={() => setMenuOpen(!menuOpen)}
                  className="flex items-center space-x-3 px-3 py-2 rounded-full hover:bg-gray-100 transition-colors"
                >
                  <div className="h-9 w-9 rounded-full bg-gradient-to-r from-purple-600 to-blue-600 text-white flex items-center justify-center font-bold">
                    {initial}
                  </div>
                  <span className="hidden sm:block font-semibold text-gray-800 max-w-[140px] truncate">{displayName}</span>
                  <ChevronDown className="h-4 w-4 text-gray-500" />
                </button>
                {menuOpen && (
                  <div className="absolute right-0 mt-2 w-48 bg-white border border-gray-200 rounded-xl shadow-lg z-50 overflow-hidden">
                    <div className="px-4 py-3 border-b border-gray-100">
                      <div className="text-sm text-gray-500">Đang đăng nhập</div>
                      <div className="text-sm font-semibold text-gray-800 truncate">{displayName}</div>
                    </div>
                    <button
                      onClick={() => { setMenuOpen(false); onGetStarted(); }}
                      className="w-full text-left px-4 py-2 text-sm hover:bg-gray-50 flex items-center space-x-2"
                    >
                      <MessageCircle className="h-4 w-4 text-gray-600" />
                      <span>Vào Chat</span>
                    </button>
                    <button
                      onClick={() => { setMenuOpen(false); setShowEditProfile(true); }}
                      className="w-full text-left px-4 py-2 text-sm hover:bg-gray-50 flex items-center space-x-2"
                    >
                      <Pencil className="h-4 w-4 text-gray-600" />
                      <span>Chỉnh sửa thông tin</span>
                    </button>
                    <button
                      onClick={handleLogout}
                      className="w-full text-left px-4 py-2 text-sm hover:bg-gray-50 flex items-center space-x-2"
                    >
                      <LogOut className="h-4 w-4 text-gray-600" />
                      <span>Đăng xuất</span>
                    </button>
                  </div>
                )}
              </div>
            ) : (
              <button
                onClick={onGetStarted}
                className="bg-gradient-to-r from-purple-600 to-blue-600 hover:from-purple-700 hover:to-blue-700 text-white px-6 py-2 rounded-full font-semibold transition-all transform hover:scale-105 shadow-lg"
              >
                Bắt đầu
              </button>
            )}
          </div>
        </div>
      </header>

      {/* Hero Section */}
      <section id="home" className="py-20 px-4 sm:px-6 lg:px-8">
        <div className="max-w-7xl mx-auto">
          <div className="grid lg:grid-cols-2 gap-12 items-center">
            {/* Left Content */}
            <div className="text-center lg:text-left">
              <h1 className="text-5xl lg:text-6xl font-bold text-gray-900 mb-6 leading-tight whitespace-nowrap">
                Trò chuyện{' '}
                <span className="bg-gradient-to-r from-purple-600 via-pink-600 to-blue-600 bg-clip-text text-transparent animate-pulse">
                  ẩn danh
                </span>{' '}
                với mọi người
              </h1>
              
              <div className="text-xl text-gray-600 mb-8 leading-relaxed">
                <p className="mb-4 whitespace-nowrap">
                  Bạn muốn chia sẻ với mọi người nhưng lại ngần ngại ?
                </p>
                <p className="mb-4">
                  Đừng lo lắng, <strong className="text-purple-600">ẨN DANH</strong> đáp ứng đủ cho bạn tiêu chí 3 KHÔNG:
                </p>
                <div className="space-y-3">
                  <div className="flex items-center space-x-3">
                    <UserX className="h-6 w-6 text-purple-600 flex-shrink-0" />
                    <span><strong>KHÔNG</strong> xác thực tài khoản</span>
                  </div>
                  <div className="flex items-center space-x-3">
                    <EyeOff className="h-6 w-6 text-purple-600 flex-shrink-0" />
                    <span><strong>KHÔNG</strong> yêu cầu thông tin cá nhân</span>
                  </div>
                  <div className="flex items-center space-x-3">
                    <X className="h-6 w-6 text-purple-600 flex-shrink-0" />
                    <span><strong>KHÔNG</strong> lưu trữ nội dung trò chuyện</span>
                  </div>
                </div>
              </div>

              {/* CTA Button */}
              <div className="flex justify-center lg:justify-start">
                <button
                  onClick={onGetStarted}
                  className="bg-gradient-to-r from-purple-600 to-blue-600 hover:from-purple-700 hover:to-blue-700 text-white px-12 py-4 rounded-full font-semibold text-lg transition-all transform hover:scale-105 shadow-xl flex items-center justify-center space-x-2"
                >
                  <MessageCircle className="h-6 w-6" />
                  <span>Bắt đầu Chat</span>
                </button>
              </div>
            </div>

            {/* Right Content - Chat Preview */}
            <div className="relative">
              <div className="grid grid-cols-2 gap-4">
                {/* Chat Avatars */}
                <div className="space-y-4">
                  <div className="bg-white rounded-2xl p-4 shadow-lg border border-purple-100 transform hover:scale-105 transition-transform">
                    <div className="flex items-center space-x-3">
                      <div className="h-12 w-12 bg-gradient-to-r from-pink-500 to-rose-500 rounded-full flex items-center justify-center text-white font-bold">
                        A
                      </div>
                      <div>
                        <div className="h-2 bg-gray-200 rounded w-16 mb-1"></div>
                        <div className="h-2 bg-gray-100 rounded w-12"></div>
                      </div>
                    </div>
                  </div>
                  
                  <div className="bg-white rounded-2xl p-4 shadow-lg border border-blue-100 transform hover:scale-105 transition-transform">
                    <div className="flex items-center space-x-3">
                      <div className="h-12 w-12 bg-gradient-to-r from-blue-500 to-cyan-500 rounded-full flex items-center justify-center text-white font-bold">
                        B
                      </div>
                      <div>
                        <div className="h-2 bg-gray-200 rounded w-20 mb-1"></div>
                        <div className="h-2 bg-gray-100 rounded w-16"></div>
                      </div>
                    </div>
                  </div>
                </div>

                <div className="space-y-4 mt-8">
                  <div className="bg-white rounded-2xl p-4 shadow-lg border border-green-100 transform hover:scale-105 transition-transform">
                    <div className="flex items-center space-x-3">
                      <div className="h-12 w-12 bg-gradient-to-r from-green-500 to-emerald-500 rounded-full flex items-center justify-center text-white font-bold">
                        C
                      </div>
                      <div>
                        <div className="h-2 bg-gray-200 rounded w-14 mb-1"></div>
                        <div className="h-2 bg-gray-100 rounded w-18"></div>
                      </div>
                    </div>
                  </div>
                  
                  <div className="bg-white rounded-2xl p-4 shadow-lg border border-orange-100 transform hover:scale-105 transition-transform">
                    <div className="flex items-center space-x-3">
                      <div className="h-12 w-12 bg-gradient-to-r from-orange-500 to-red-500 rounded-full flex items-center justify-center text-white font-bold">
                        D
                      </div>
                      <div>
                        <div className="h-2 bg-gray-200 rounded w-12 mb-1"></div>
                        <div className="h-2 bg-gray-100 rounded w-20"></div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              {/* Floating Elements */}
              <div className="absolute -top-4 -right-4 h-16 w-16 bg-gradient-to-r from-purple-500 to-pink-500 rounded-full flex items-center justify-center text-white shadow-lg animate-bounce">
                <Heart className="h-8 w-8" />
              </div>
              
              <div className="absolute -bottom-4 -left-4 h-12 w-12 bg-gradient-to-r from-blue-500 to-cyan-500 rounded-full flex items-center justify-center text-white shadow-lg animate-pulse">
                <Globe className="h-6 w-6" />
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Features Section */}
      <section id="features" className="py-20 bg-white">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16">
            <h2 className="text-4xl font-bold text-gray-900 mb-4">
              Tính năng nổi bật
            </h2>
            <p className="text-xl text-gray-600 max-w-3xl mx-auto">
              ChatMix mang đến trải nghiệm chat hiện đại, an toàn và thú vị để bạn kết nối với mọi người
            </p>
          </div>

          <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-8">
            {features.map((feature, index) => (
              <div
                key={index}
                className="bg-gradient-to-br from-white to-gray-50 rounded-2xl p-6 border border-gray-100 hover:shadow-xl transition-all transform hover:-translate-y-2 group"
              >
                <div className="h-16 w-16 bg-gradient-to-r from-purple-600 to-blue-600 rounded-2xl flex items-center justify-center text-white mb-4 group-hover:scale-110 transition-transform">
                  {feature.icon}
                </div>
                <h3 className="text-xl font-semibold text-gray-900 mb-2">
                  {feature.title}
                </h3>
                <p className="text-gray-600">
                  {feature.description}
                </p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Stats Section */}
      <section className="py-20 bg-gradient-to-r from-purple-600 to-blue-600 text-white">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16">
            <h2 className="text-4xl font-bold mb-4">
              Tham gia cộng đồng ChatMix
            </h2>
            <p className="text-xl opacity-90">
              Hàng ngàn người dùng đang kết nối qua ChatMix và tạo ra những tình bạn mới mỗi ngày
            </p>
          </div>

          <div className="grid md:grid-cols-3 gap-8 text-center">
            <div className="bg-white/10 backdrop-blur-md rounded-2xl p-8">
              <div className="text-4xl font-bold mb-2">1,000+</div>
              <div className="text-lg opacity-90">Người dùng hoạt động</div>
            </div>
            <div className="bg-white/10 backdrop-blur-md rounded-2xl p-8">
              <div className="text-4xl font-bold mb-2">50K+</div>
              <div className="text-lg opacity-90">Tin nhắn gửi</div>
            </div>
            <div className="bg-white/10 backdrop-blur-md rounded-2xl p-8">
              <div className="text-4xl font-bold mb-2">24/7</div>
              <div className="text-lg opacity-90">Hỗ trợ</div>
            </div>
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section className="py-20 bg-gradient-to-br from-purple-900 via-purple-800 to-blue-900 text-white">
        <div className="max-w-4xl mx-auto text-center px-4 sm:px-6 lg:px-8">
          <h2 className="text-4xl lg:text-5xl font-bold mb-6">
            Sẵn sàng bắt đầu cuộc trò chuyện?
          </h2>
          <p className="text-xl mb-8 opacity-90">
            Tham gia ChatMix ngay hôm nay và khám phá thế giới kết nối mới
          </p>
          
          <button
            onClick={onGetStarted}
            className="bg-white text-purple-900 px-12 py-4 rounded-full font-bold text-lg hover:bg-gray-100 transition-all transform hover:scale-105 shadow-2xl inline-flex items-center space-x-2"
          >
            <span>Bắt đầu ngay</span>
            <ArrowRight className="h-5 w-5" />
          </button>
        </div>
      </section>

      {/* Footer */}
      <footer className="bg-gray-900 text-white py-12">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="grid md:grid-cols-4 gap-8">
            <div>
              <div className="flex items-center space-x-2 mb-4">
                <div className="h-8 w-8 bg-gradient-to-r from-purple-600 to-blue-600 rounded-full flex items-center justify-center">
                  <MessageCircle className="h-5 w-5 text-white" />
                </div>
                <span className="text-xl font-bold">ChatMix</span>
              </div>
              <div className="text-gray-400">
                <p className="mb-3">
                  Bạn muốn chia sẻ với mọi người nhưng lại ngần ngại vì sợ họ biết bạn là ai?
                </p>
              </div>
            </div>
            
            <div>
              <h3 className="text-lg font-semibold mb-4">Trung tâm trợ giúp</h3>
              <ul className="space-y-2 text-gray-400">
                <li><span className="cursor-default">Facebook</span></li>
                <li><span className="cursor-default">Instagram</span></li>
                <li><span className="cursor-default">Tiktok</span></li>
              </ul>
            </div>
            
            <div>
              <h3 className="text-lg font-semibold mb-4">Điều khoản sử dụng</h3>
              <ul className="space-y-2 text-gray-400">
                <li><span className="cursor-default">Facebook</span></li>
                <li><span className="cursor-default">Instagram</span></li>
                <li><span className="cursor-default">Tiktok</span></li>
              </ul>
            </div>

            <div>
              <h3 className="text-lg font-semibold mb-4">Liên hệ</h3>
              <ul className="space-y-2 text-gray-400">
                <li><span className="cursor-default">Facebook</span></li>
                <li><span className="cursor-default">Instagram</span></li>
                <li><span className="cursor-default">Tiktok</span></li>
              </ul>
            </div>
          </div>
          
          
          <div className="border-t border-gray-800 mt-12 pt-8 text-center text-gray-400">
            <p>&copy; 2025 ChatMix.</p>
          </div>
        </div>
      </footer>

      {/* Edit Profile Modal */}
      <EditProfileModal 
        isOpen={showEditProfile} 
        onClose={() => setShowEditProfile(false)} 
        onSave={() => {
          // Có thể refresh user data ở đây nếu cần
        }}
      />
    </div>
  );
};

export default HomePage;
