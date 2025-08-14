import React, { useState, useEffect, useCallback } from 'react';
import { X, Save, Calendar, Heart, Loader2 } from 'lucide-react';
import CONFIG from '../config/api';
import { useAuth } from '../contexts/AuthContext';
import toast from 'react-hot-toast';

const EditProfileModal = ({ isOpen, onClose, onSave }) => {
  const { user } = useAuth();
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  // Remove unused profile state
  const [formData, setFormData] = useState({
    bio: '',
    age: '',
    gender: ''
  });

  const fetchProfile = useCallback(async () => {
    setLoading(true);
    try {
      const response = await fetch(`${CONFIG.API_BASE_URL}/users/${encodeURIComponent(user.username)}`);
      if (response.ok) {
        const data = await response.json();
        // Profile data loaded successfully
        setFormData({
          bio: data.bio || '',
          age: data.age || '',
          gender: data.gender || ''
        });
      } else {
        throw new Error('Không thể tải thông tin profile');
      }
    } catch (error) {
      console.error('Error fetching profile:', error);
      toast.error('Không thể tải thông tin profile');
    } finally {
      setLoading(false);
    }
  }, [user?.username]);

  useEffect(() => {
    if (isOpen && user?.username) {
      fetchProfile();
    }
  }, [isOpen, user?.username, fetchProfile]);

  const handleInputChange = (field, value) => {
    setFormData(prev => ({
      ...prev,
      [field]: value
    }));
  };

  const handleSave = async () => {
    setSaving(true);
    try {
      const token = localStorage.getItem('access_token');
      if (!token) {
        throw new Error('Không tìm thấy token đăng nhập');
      }

      const response = await fetch(`${CONFIG.API_BASE_URL}/auth/profile`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify({
          age: formData.age ? parseInt(formData.age) : null,
          gender: formData.gender,
          bio: formData.bio
        })
      });

      if (response.ok) {
        toast.success('Cập nhật thông tin thành công!');
        onSave && onSave();
        onClose();
      } else {
        const errorData = await response.json();
        throw new Error(errorData.message || 'Không thể cập nhật thông tin');
      }
    } catch (error) {
      console.error('Error updating profile:', error);
      toast.error(error.message || 'Lỗi khi cập nhật thông tin');
    } finally {
      setSaving(false);
    }
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
      <div className="bg-white rounded-3xl shadow-2xl max-w-lg w-full max-h-[90vh] overflow-hidden">
        
        {/* Header với background gradient */}
        <div className="relative bg-gradient-to-br from-blue-500 via-purple-500 to-pink-500 p-6 text-white">
          {/* Close button */}
          <button
            onClick={onClose}
            className="absolute top-4 right-4 p-2 hover:bg-white/20 rounded-full transition-colors"
          >
            <X className="h-5 w-5" />
          </button>

          {/* Header */}
          <div className="text-center">
            <div className="w-20 h-20 bg-white/20 rounded-full flex items-center justify-center mx-auto mb-4 text-3xl backdrop-blur-sm border border-white/30">
              {getGenderIcon(formData.gender)}
            </div>
            <h2 className="text-2xl font-bold mb-1">Chỉnh sửa thông tin</h2>
            <p className="text-white/80 text-sm">@{user?.username}</p>
          </div>
        </div>

        {/* Content */}
        <div className="p-6 max-h-[60vh] overflow-y-auto">
          {loading ? (
            <div className="flex items-center justify-center py-8">
              <Loader2 className="h-8 w-8 animate-spin text-blue-500" />
              <span className="ml-2 text-gray-600">Đang tải...</span>
            </div>
          ) : (
            <div className="space-y-6">
              
              {/* Nickname removed - using username only */}

              {/* Bio */}
              <div>
                <label className="flex items-center mb-3 text-sm font-semibold text-gray-700">
                  <Heart className="h-4 w-4 mr-2 text-pink-500" />
                  Giới thiệu bản thân
                </label>
                <textarea
                  value={formData.bio}
                  onChange={(e) => handleInputChange('bio', e.target.value)}
                  placeholder="Viết vài dòng về bản thân..."
                  rows={4}
                  className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all resize-none"
                  maxLength={500}
                />
                <div className="text-right text-xs text-gray-400 mt-1">
                  {formData.bio.length}/500 ký tự
                </div>
              </div>

              {/* Age */}
              <div>
                <label className="flex items-center mb-3 text-sm font-semibold text-gray-700">
                  <Calendar className="h-4 w-4 mr-2 text-green-500" />
                  Tuổi
                </label>
                <input
                  type="number"
                  value={formData.age}
                  onChange={(e) => handleInputChange('age', e.target.value)}
                  placeholder="Nhập tuổi..."
                  min="13"
                  max="100"
                  className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all"
                />
              </div>

              {/* Gender */}
              <div>
                <label className="flex items-center mb-3 text-sm font-semibold text-gray-700">
                  <span className="text-lg mr-2">👤</span>
                  Giới tính
                </label>
                <select
                  value={formData.gender}
                  onChange={(e) => handleInputChange('gender', e.target.value)}
                  className="w-full px-4 py-3 border border-gray-200 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all"
                >
                  <option value="">Chọn giới tính</option>
                  <option value="male">Nam</option>
                  <option value="female">Nữ</option>
                  <option value="other">Khác</option>
                  <option value="private">Riêng tư</option>
                </select>
              </div>
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="px-6 pb-6 flex space-x-3">
          <button
            onClick={onClose}
            className="flex-1 px-6 py-3 border border-gray-200 text-gray-700 rounded-xl hover:bg-gray-50 transition-colors font-medium"
          >
            Hủy
          </button>
          <button
            onClick={handleSave}
            disabled={saving || loading}
            className="flex-1 bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700 text-white px-6 py-3 rounded-xl transition-all duration-200 shadow-lg hover:shadow-xl font-medium flex items-center justify-center disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {saving ? (
              <>
                <Loader2 className="h-4 w-4 animate-spin mr-2" />
                Đang lưu...
              </>
            ) : (
              <>
                <Save className="h-4 w-4 mr-2" />
                Lưu thay đổi
              </>
            )}
          </button>
        </div>
      </div>
    </div>
  );
};

export default EditProfileModal;
