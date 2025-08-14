import React from 'react';

const ChatHello = ({ onBack }) => {
  return (
    <div className="min-h-screen bg-gradient-to-br from-purple-50 via-white to-blue-50 flex items-center justify-center p-6">
      <div className="bg-white rounded-3xl shadow-xl border border-gray-100 p-10 text-center max-w-lg w-full">
        <h1 className="text-4xl font-extrabold text-gray-900 mb-4">Hello World 👋</h1>
        <p className="text-gray-600 mb-8">Bạn đã đăng nhập và vào khu vực chat mẫu.</p>
        <button
          onClick={onBack}
          className="inline-flex items-center justify-center px-6 py-3 rounded-xl bg-gradient-to-r from-purple-600 to-blue-600 text-white font-semibold shadow hover:from-purple-700 hover:to-blue-700 transition-colors"
        >
          Quay lại
        </button>
      </div>
    </div>
  );
};

export default ChatHello;


