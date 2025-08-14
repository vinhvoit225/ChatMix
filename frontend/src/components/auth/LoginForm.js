import React, { useState, useEffect } from 'react';
import { Eye, EyeOff, User, Lock, RefreshCw, LogIn } from 'lucide-react';
import { useAuth } from '../../contexts/AuthContext';
import authService from '../../services/authService';

const LoginForm = ({ onSwitchToRegister }) => {
  const { login, isLoading, error } = useAuth();
  
  const [formData, setFormData] = useState({
    username: '',
    password: '',
    captcha: '',
  });
  
  const [showPassword, setShowPassword] = useState(false);
  const [captchaData, setCaptchaData] = useState(null);
  const [captchaLoading, setCaptchaLoading] = useState(false);
  const [formErrors, setFormErrors] = useState({});

  // Load captcha on component mount
  useEffect(() => {
    loadCaptcha();
  }, []);

  const loadCaptcha = async () => {
    setCaptchaLoading(true);
    try {
      const captcha = await authService.generateCaptcha();
      setCaptchaData(captcha);
      setFormData(prev => ({ ...prev, captcha: '' }));
    } catch (error) {
      console.error('Failed to load captcha:', error);
    } finally {
      setCaptchaLoading(false);
    }
  };

  const handleInputChange = (e) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value,
    }));
    
    // Clear field error when user starts typing
    if (formErrors[name]) {
      setFormErrors(prev => ({
        ...prev,
        [name]: '',
      }));
    }
  };

  const validateForm = () => {
    const errors = {};

    if (!formData.username.trim()) {
      errors.username = 'Tên đăng nhập không được để trống';
    }

    if (!formData.password) {
      errors.password = 'Mật khẩu không được để trống';
    }

    if (!formData.captcha.trim()) {
      errors.captcha = 'Vui lòng nhập kết quả phép tính';
    }

    setFormErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    if (!validateForm()) return;

    try {
      const loginData = {
        username: formData.username.trim(),
        password: formData.password,
        captcha: captchaData?.challenge_id || '',
        captcha_answer: formData.captcha.trim(),
      };

      await login(loginData);
    } catch (error) {
      // Error is handled by AuthContext
      // Reload captcha after failed attempt
      loadCaptcha();
    }
  };

  return (
    <div className="w-full max-w-md mx-auto">
      <div className="glass-effect rounded-3xl p-8 shadow-2xl animate-fade-in">
        {/* Header */}
        <div className="text-center mb-8">
          <div className="mx-auto h-16 w-16 bg-gradient-to-r from-blue-500 to-purple-600 rounded-full flex items-center justify-center mb-4 shadow-lg">
            <LogIn className="h-8 w-8 text-white" />
          </div>
          <h2 className="text-3xl font-bold text-gray-800 mb-2">Đăng nhập</h2>
          <p className="text-gray-600">Chào mừng bạn quay lại ChatMix! 👋</p>
        </div>

        {/* Error Message */}
        {error && (
          <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-xl text-red-700 text-sm animate-slide-up">
            {error}
          </div>
        )}

        {/* Login Form */}
        <form onSubmit={handleSubmit} className="space-y-6">
          {/* Username Field */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Tên đăng nhập hoặc Email
            </label>
            <div className="relative">
              <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                <User className="h-5 w-5 text-gray-400" />
              </div>
              <input
                type="text"
                name="username"
                value={formData.username}
                onChange={handleInputChange}
                className={`w-full pl-10 pr-4 py-3 border rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-colors ${
                  formErrors.username ? 'border-red-300 bg-red-50' : 'border-gray-300'
                }`}
                placeholder="Nhập tên đăng nhập hoặc email..."
                disabled={isLoading}
              />
            </div>
            {formErrors.username && (
              <p className="mt-1 text-sm text-red-600">{formErrors.username}</p>
            )}
          </div>

          {/* Password Field */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Mật khẩu
            </label>
            <div className="relative">
              <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                <Lock className="h-5 w-5 text-gray-400" />
              </div>
              <input
                type={showPassword ? 'text' : 'password'}
                name="password"
                value={formData.password}
                onChange={handleInputChange}
                className={`w-full pl-10 pr-12 py-3 border rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-colors ${
                  formErrors.password ? 'border-red-300 bg-red-50' : 'border-gray-300'
                }`}
                placeholder="Nhập mật khẩu..."
                disabled={isLoading}
              />
              <button
                type="button"
                onClick={() => setShowPassword(!showPassword)}
                className="absolute inset-y-0 right-0 pr-3 flex items-center text-gray-400 hover:text-gray-600"
                disabled={isLoading}
              >
                {showPassword ? <EyeOff className="h-5 w-5" /> : <Eye className="h-5 w-5" />}
              </button>
            </div>
            {formErrors.password && (
              <p className="mt-1 text-sm text-red-600">{formErrors.password}</p>
            )}
          </div>

          {/* Captcha Field */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Xác thực
            </label>
            <div className="flex space-x-3">
              <div className="flex-1">
                <div className="flex items-center space-x-3 mb-2">
                  <div className="bg-gray-100 border-2 border-dashed border-gray-300 rounded-lg px-4 py-3 font-mono text-lg font-semibold text-gray-700 min-h-[50px] flex items-center justify-center">
                    {captchaLoading ? (
                      <div className="animate-pulse">Đang tải...</div>
                    ) : (
                      captchaData?.challenge || 'Chưa tải'
                    )}
                  </div>
                  <button
                    type="button"
                    onClick={loadCaptcha}
                    disabled={captchaLoading}
                    className="p-2 text-gray-500 hover:text-gray-700 hover:bg-gray-100 rounded-lg transition-colors"
                    title="Tải lại captcha"
                  >
                    <RefreshCw className={`h-5 w-5 ${captchaLoading ? 'animate-spin' : ''}`} />
                  </button>
                </div>
                <input
                  type="text"
                  name="captcha"
                  value={formData.captcha}
                  onChange={handleInputChange}
                  className={`w-full px-4 py-3 border rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-colors ${
                    formErrors.captcha ? 'border-red-300 bg-red-50' : 'border-gray-300'
                  }`}
                  placeholder="Nhập kết quả..."
                  disabled={isLoading}
                />
                {formErrors.captcha && (
                  <p className="mt-1 text-sm text-red-600">{formErrors.captcha}</p>
                )}
              </div>
            </div>
          </div>

          {/* Submit Button */}
          <button
            type="submit"
            disabled={isLoading}
            className="w-full bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700 disabled:from-gray-400 disabled:to-gray-500 text-white font-semibold py-3 px-6 rounded-xl transition-all duration-200 transform hover:scale-105 disabled:hover:scale-100 disabled:cursor-not-allowed shadow-lg"
          >
            {isLoading ? (
              <div className="flex items-center justify-center">
                <RefreshCw className="animate-spin h-5 w-5 mr-2" />
                Đang đăng nhập...
              </div>
            ) : (
              <div className="flex items-center justify-center">
                <LogIn className="h-5 w-5 mr-2" />
                Đăng nhập
              </div>
            )}
          </button>
        </form>

        {/* Switch to Register */}
        <div className="mt-8 text-center">
          <p className="text-gray-600">
            Chưa có tài khoản?{' '}
            <button
              onClick={onSwitchToRegister}
              className="text-blue-600 hover:text-blue-700 font-semibold hover:underline transition-colors"
              disabled={isLoading}
            >
              Đăng ký ngay
            </button>
          </p>
        </div>
      </div>
    </div>
  );
};

export default LoginForm;
