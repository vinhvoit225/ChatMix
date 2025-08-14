import React, { useState, useEffect, useCallback } from 'react';
import { X, User, Calendar, MessageCircle, Heart } from 'lucide-react';
import CONFIG from '../config/api';

const ProfileModal = ({ username, isOpen, onClose }) => {
  const [profile, setProfile] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const fetchProfile = useCallback(async () => {
    setLoading(true);
    setError('');
    
    try {
      const response = await fetch(`${CONFIG.API_BASE_URL}/users/${encodeURIComponent(username)}`);
      
      if (!response.ok) {
        throw new Error('Không thể tải thông tin người dùng');
      }
      
      const data = await response.json();
      setProfile(data);
    } catch (err) {
      setError(err.message);
      console.error('Error fetching profile:', err);
    } finally {
      setLoading(false);
    }
  }, [username]);

  useEffect(() => {
    if (isOpen && username) {
      fetchProfile();
    }
  }, [isOpen, username, fetchProfile]);

  // fetchProfile function moved above as useCallback

  const formatGender = (gender) => {
    switch (gender) {
      case 'male': return 'Nam';
      case 'female': return 'Nữ';
      case 'other': return 'Khác';
      case 'private': return 'Riêng tư';
      default: return 'Chưa cập nhật';
    }
  };

  const formatJoinDate = (dateString) => {
    const date = new Date(dateString);
    return date.toLocaleDateString('vi-VN', { 
      year: 'numeric', 
      month: 'long', 
      day: 'numeric' 
    });
  };

  const getGenderIcon = (gender) => {
    switch (gender) {
      case 'male': return '👨';
      case 'female': return '👩';
      case 'other': return '🧑';
      default: return '👤';
    }
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-3xl shadow-2xl max-w-md w-full max-h-[90vh] overflow-hidden">
        
        {/* Header với background gradient */}
        <div className="relative bg-gradient-to-br from-purple-500 via-pink-500 to-rose-500 p-6 text-white">
          {/* Close button */}
          <button
            onClick={onClose}
            className="absolute top-4 right-4 p-2 hover:bg-white/20 rounded-full transition-colors"
          >
            <X className="h-5 w-5" />
          </button>

          {/* Profile header */}
          <div className="text-center">
            <div className="w-20 h-20 bg-white/20 rounded-full flex items-center justify-center mx-auto mb-4 text-3xl backdrop-blur-sm border border-white/30">
              {profile ? getGenderIcon(profile.gender) : '👤'}
            </div>
            
            {loading ? (
              <div className="animate-pulse">
                <div className="h-6 bg-white/30 rounded w-32 mx-auto mb-2"></div>
                <div className="h-4 bg-white/20 rounded w-24 mx-auto"></div>
              </div>
            ) : error ? (
              <div className="text-red-200">
                <p className="font-semibold">Lỗi tải dữ liệu</p>
                <p className="text-sm opacity-90">{error}</p>
              </div>
            ) : profile ? (
              <>
                <h2 className="text-2xl font-bold mb-1">{profile.username}</h2>
                <p className="text-white/80 text-sm">@{profile.username}</p>
                <div className="flex items-center justify-center mt-2">
                  <div className={`w-3 h-3 rounded-full mr-2 ${
                    profile.is_online ? 'bg-green-400 animate-pulse' : 'bg-gray-400'
                  }`}></div>
                  <span className="text-sm">
                    {profile.is_online ? 'Đang trực tuyến' : 'Không trực tuyến'}
                  </span>
                </div>
              </>
            ) : null}
          </div>
        </div>

        {/* Content */}
        <div className="p-6">
          {loading ? (
            <div className="space-y-4">
              {[1, 2, 3].map(i => (
                <div key={i} className="animate-pulse">
                  <div className="h-4 bg-gray-200 rounded w-3/4 mb-2"></div>
                  <div className="h-3 bg-gray-100 rounded w-1/2"></div>
                </div>
              ))}
            </div>
          ) : error ? (
            <div className="text-center py-8">
              <MessageCircle className="h-12 w-12 text-gray-300 mx-auto mb-4" />
              <p className="text-gray-500">Không thể tải thông tin người dùng</p>
              <button
                onClick={fetchProfile}
                className="mt-4 px-4 py-2 bg-purple-600 text-white rounded-lg hover:bg-purple-700 transition-colors"
              >
                Thử lại
              </button>
            </div>
          ) : profile ? (
            <div className="space-y-6">
              
              {/* Bio */}
              {profile.bio && (
                <div>
                  <div className="flex items-center mb-3">
                    <Heart className="h-5 w-5 text-pink-500 mr-2" />
                    <h3 className="font-semibold text-gray-800">Giới thiệu</h3>
                  </div>
                  <p className="text-gray-600 bg-gray-50 rounded-xl p-4 leading-relaxed">
                    {profile.bio}
                  </p>
                </div>
              )}

              {/* Details */}
              <div>
                <div className="flex items-center mb-3">
                  <User className="h-5 w-5 text-blue-500 mr-2" />
                  <h3 className="font-semibold text-gray-800">Thông tin</h3>
                </div>
                
                <div className="space-y-3">
                  {/* Age */}
                  {profile.age && (
                    <div className="flex items-center justify-between py-2 px-4 bg-gray-50 rounded-xl">
                      <span className="text-gray-600">Tuổi</span>
                      <span className="font-medium text-gray-800">{profile.age} tuổi</span>
                    </div>
                  )}
                  
                  {/* Gender */}
                  {profile.gender && profile.gender !== 'private' && (
                    <div className="flex items-center justify-between py-2 px-4 bg-gray-50 rounded-xl">
                      <span className="text-gray-600">Giới tính</span>
                      <span className="font-medium text-gray-800">{formatGender(profile.gender)}</span>
                    </div>
                  )}
                  
                  {/* Join date */}
                  {profile.joined_at && (
                    <div className="flex items-center justify-between py-2 px-4 bg-gray-50 rounded-xl">
                      <span className="text-gray-600 flex items-center">
                        <Calendar className="h-4 w-4 mr-1" />
                        Tham gia
                      </span>
                      <span className="font-medium text-gray-800">{formatJoinDate(profile.joined_at)}</span>
                    </div>
                  )}
                </div>
              </div>

              {/* Verification badge */}
              {profile.is_verified && (
                <div className="flex items-center justify-center py-3 px-4 bg-green-50 rounded-xl border border-green-200">
                  <div className="w-2 h-2 bg-green-500 rounded-full mr-2"></div>
                  <span className="text-green-700 font-medium">Tài khoản đã xác thực</span>
                </div>
              )}
            </div>
          ) : null}
        </div>

        {/* Footer */}
        <div className="px-6 pb-6">
          <button
            onClick={onClose}
            className="w-full bg-gradient-to-r from-purple-600 to-pink-600 hover:from-purple-700 hover:to-pink-700 text-white font-semibold py-3 px-6 rounded-xl transition-all duration-200 shadow-lg hover:shadow-xl"
          >
            Đóng
          </button>
        </div>
      </div>
    </div>
  );
};

export default ProfileModal;