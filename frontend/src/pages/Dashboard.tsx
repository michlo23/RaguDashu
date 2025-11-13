import React from 'react';
import { useAuthStore } from '../stores/authStore';
import { FileText, MessageSquare, Search, Key, Database } from 'lucide-react';
import { Link } from 'react-router-dom';

export default function Dashboard() {
  const { user } = useAuthStore();

  const features = [
    {
      title: 'Documents',
      description: 'Upload and manage your documents for RAG',
      icon: FileText,
      link: '/documents',
      color: 'bg-blue-500',
    },
    {
      title: 'AI Chat',
      description: 'Chat with AI using your documents as context',
      icon: MessageSquare,
      link: '/chat',
      color: 'bg-green-500',
    },
    {
      title: 'Search',
      description: 'Semantic search across your knowledge base',
      icon: Search,
      link: '/search',
      color: 'bg-purple-500',
    },
    {
      title: 'Credentials',
      description: 'Manage your API keys (OpenAI, Pinecone)',
      icon: Key,
      link: '/credentials',
      color: 'bg-orange-500',
    },
  ];

  return (
    <div className="min-h-screen bg-gray-50">
      <nav className="bg-white shadow">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16">
            <div className="flex items-center">
              <Database className="h-8 w-8 text-indigo-600" />
              <span className="ml-2 text-xl font-bold text-gray-900">RAG Dashboard</span>
            </div>
            <div className="flex items-center space-x-4">
              <span className="text-sm text-gray-700">{user?.email}</span>
              <button
                onClick={() => useAuthStore.getState().logout()}
                className="text-sm text-gray-700 hover:text-gray-900"
              >
                Logout
              </button>
            </div>
          </div>
        </div>
      </nav>

      <div className="max-w-7xl mx-auto py-12 px-4 sm:px-6 lg:px-8">
        <div className="text-center mb-12">
          <h1 className="text-4xl font-extrabold text-gray-900 sm:text-5xl">
            Welcome to RAG Dashboard
          </h1>
          <p className="mt-4 text-xl text-gray-600">
            Your AI-powered knowledge management system
          </p>
        </div>

        <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-4">
          {features.map((feature) => (
            <Link
              key={feature.title}
              to={feature.link}
              className="group relative bg-white rounded-lg shadow-sm hover:shadow-md transition-shadow p-6"
            >
              <div className={`inline-flex p-3 rounded-lg ${feature.color} text-white mb-4`}>
                <feature.icon className="h-6 w-6" />
              </div>
              <h3 className="text-lg font-medium text-gray-900 group-hover:text-indigo-600">
                {feature.title}
              </h3>
              <p className="mt-2 text-sm text-gray-500">{feature.description}</p>
            </Link>
          ))}
        </div>

        <div className="mt-12 bg-white rounded-lg shadow p-6">
          <h2 className="text-2xl font-bold text-gray-900 mb-4">Getting Started</h2>
          <ol className="list-decimal list-inside space-y-2 text-gray-700">
            <li>Add your OpenAI and Pinecone API keys in <Link to="/credentials" className="text-indigo-600 hover:underline">Credentials</Link></li>
            <li>Upload documents to build your knowledge base</li>
            <li>Configure a chat assistant with your preferred settings</li>
            <li>Start chatting with AI that has access to your documents!</li>
          </ol>
        </div>
      </div>
    </div>
  );
}
