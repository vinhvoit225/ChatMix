import React, { createContext, useContext, useReducer, useEffect } from 'react';
import authService from '../services/authService';
import toast from 'react-hot-toast';

// Initial state
const initialState = {
  user: null,
  isAuthenticated: false,
  isLoading: true,
  error: null,
};

// Action types
const AUTH_ACTIONS = {
  LOGIN_START: 'LOGIN_START',
  LOGIN_SUCCESS: 'LOGIN_SUCCESS',
  LOGIN_FAILURE: 'LOGIN_FAILURE',
  LOGOUT: 'LOGOUT',
  SET_USER: 'SET_USER',
  SET_LOADING: 'SET_LOADING',
  CLEAR_ERROR: 'CLEAR_ERROR',
};

// Reducer
function authReducer(state, action) {
  switch (action.type) {
    case AUTH_ACTIONS.LOGIN_START:
      return {
        ...state,
        isLoading: true,
        error: null,
      };

    case AUTH_ACTIONS.LOGIN_SUCCESS:
      return {
        ...state,
        user: action.payload.user,
        isAuthenticated: true,
        isLoading: false,
        error: null,
      };

    case AUTH_ACTIONS.LOGIN_FAILURE:
      return {
        ...state,
        user: null,
        isAuthenticated: false,
        isLoading: false,
        error: action.payload.error,
      };

    case AUTH_ACTIONS.LOGOUT:
      return {
        ...state,
        user: null,
        isAuthenticated: false,
        isLoading: false,
        error: null,
      };

    case AUTH_ACTIONS.SET_USER:
      return {
        ...state,
        user: action.payload.user,
        isAuthenticated: !!action.payload.user,
        isLoading: false,
      };

    case AUTH_ACTIONS.SET_LOADING:
      return {
        ...state,
        isLoading: action.payload.isLoading,
      };

    case AUTH_ACTIONS.CLEAR_ERROR:
      return {
        ...state,
        error: null,
      };

    default:
      return state;
  }
}

// Create context
const AuthContext = createContext();

// Auth Provider Component
export function AuthProvider({ children }) {
  const [state, dispatch] = useReducer(authReducer, initialState);

  // Initialize auth state on app load
  useEffect(() => {
    initializeAuth();

    // Listen for auth logout events from API interceptors
    const handleAuthLogout = () => {
      dispatch({ type: AUTH_ACTIONS.LOGOUT });
    };

    window.addEventListener('auth:logout', handleAuthLogout);

    return () => {
      window.removeEventListener('auth:logout', handleAuthLogout);
    };
  }, []);

  const initializeAuth = async () => {
    dispatch({ type: AUTH_ACTIONS.SET_LOADING, payload: { isLoading: true } });

    try {
      if (authService.isAuthenticated() && !authService.isTokenExpired()) {
        // Token exists and is valid, get user profile
        const profile = await authService.getProfile();
        dispatch({
          type: AUTH_ACTIONS.SET_USER,
          payload: { user: profile },
        });
      } else if (authService.refreshToken) {
        // Try to refresh token
        try {
          await authService.refreshAccessToken();
          const profile = await authService.getProfile();
          dispatch({
            type: AUTH_ACTIONS.SET_USER,
            payload: { user: profile },
          });
        } catch (error) {
          // Refresh failed, clear tokens
          authService.clearTokens();
          dispatch({ type: AUTH_ACTIONS.LOGOUT });
        }
      } else {
        // No valid authentication
        dispatch({ type: AUTH_ACTIONS.LOGOUT });
      }
    } catch (error) {
      console.error('Auth initialization failed:', error);
      dispatch({ type: AUTH_ACTIONS.LOGOUT });
    }
  };

  const login = async (credentials) => {
    dispatch({ type: AUTH_ACTIONS.LOGIN_START });

    try {
      const response = await authService.login(credentials);
      
      dispatch({
        type: AUTH_ACTIONS.LOGIN_SUCCESS,
        payload: { user: response.user },
      });

      toast.success(`Chào mừng ${response.user.username}! 🎉`);
      return response;
    } catch (error) {
      const errorMessage = error.message || 'Đăng nhập thất bại';
      
      dispatch({
        type: AUTH_ACTIONS.LOGIN_FAILURE,
        payload: { error: errorMessage },
      });

      toast.error(errorMessage);
      throw error;
    }
  };

  const register = async (userData) => {
    dispatch({ type: AUTH_ACTIONS.LOGIN_START });

    try {
      const response = await authService.register(userData);
      
      dispatch({
        type: AUTH_ACTIONS.LOGIN_SUCCESS,
        payload: { user: response.user },
      });

      toast.success(`Đăng ký thành công! Chào mừng ${response.user.username}! 🎊`);
      return response;
    } catch (error) {
      const errorMessage = error.message || 'Đăng ký thất bại';
      
      dispatch({
        type: AUTH_ACTIONS.LOGIN_FAILURE,
        payload: { error: errorMessage },
      });

      toast.error(errorMessage);
      throw error;
    }
  };

  const logout = async () => {
    try {
      await authService.logout();
      dispatch({ type: AUTH_ACTIONS.LOGOUT });
      toast.success('Đã đăng xuất thành công! 👋');
    } catch (error) {
      console.error('Logout failed:', error);
      // Force logout even if API call fails
      authService.clearTokens();
      dispatch({ type: AUTH_ACTIONS.LOGOUT });
      toast.success('Đã đăng xuất! 👋');
    }
  };

  const updateProfile = async (profileData) => {
    try {
      const updatedUser = await authService.updateProfile(profileData);
      
      dispatch({
        type: AUTH_ACTIONS.SET_USER,
        payload: { user: updatedUser },
      });

      toast.success('Cập nhật profile thành công! ✨');
      return updatedUser;
    } catch (error) {
      const errorMessage = error.message || 'Cập nhật profile thất bại';
      toast.error(errorMessage);
      throw error;
    }
  };

  const changePassword = async (passwordData) => {
    try {
      await authService.changePassword(passwordData);
      toast.success('Đổi mật khẩu thành công! 🔐');
    } catch (error) {
      const errorMessage = error.message || 'Đổi mật khẩu thất bại';
      toast.error(errorMessage);
      throw error;
    }
  };

  const refreshUser = async () => {
    try {
      if (!authService.isAuthenticated()) return;

      await authService.ensureValidToken();
      const profile = await authService.getProfile();
      
      dispatch({
        type: AUTH_ACTIONS.SET_USER,
        payload: { user: profile },
      });

      return profile;
    } catch (error) {
      console.error('Failed to refresh user:', error);
      // If refresh fails, logout user
      logout();
    }
  };

  const clearError = () => {
    dispatch({ type: AUTH_ACTIONS.CLEAR_ERROR });
  };

  // Context value
  const value = {
    ...state,
    login,
    register,
    logout,
    updateProfile,
    changePassword,
    refreshUser,
    clearError,
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
}

// Custom hook to use auth context
export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}

export default AuthContext;
