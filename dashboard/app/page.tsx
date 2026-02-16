"use client";
import { useState, useEffect } from 'react';
import { api, ExperimentResponse } from '@/lib/api';

export default function Dashboard() {
  const [users, setUsers] = useState<{user_id: string, orders: number}[]>([]);
  const [selectedUser, setSelectedUser] = useState<string | null>(null);
  const [exp, setExp] = useState<ExperimentResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [newUserId, setNewUserId] = useState('');

  // Load users on mount
  const refreshUsers = async () => {
    const data = await api.getUsers();
    setUsers(data);
  };

  useEffect(() => { refreshUsers(); }, []);

  // Fetch experiments when user changes
  useEffect(() => {
    if (selectedUser) {
      api.getExperiments(selectedUser).then(setExp);
    }
  }, [selectedUser]);

  const handleSimulateOrders = async () => {
    if (!selectedUser) return;
    setLoading(true);
    await api.placeOrders(selectedUser, 30);
    // Give the worker/cron a second to process
    setTimeout(async () => {
      await refreshUsers();
      const newExp = await api.getExperiments(selectedUser);
      setExp(newExp);
      setLoading(false);
    }, 2000);
  };

  const handleSync = async () => {
    setLoading(true);
    await fetch("http://localhost:8080/evaluate", { method: 'POST' });
    const newExp = await api.getExperiments(selectedUser!);
    setExp(newExp);
    setLoading(false);
  };

  const handleCreateUser = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newUserId) return;
    
    await api.createUser(newUserId);
    setNewUserId('');
    await refreshUsers(); // Refresh the list
  };

  return (
    <main className="flex min-h-screen bg-gray-950 text-white p-8 font-sans">
      {/* Sidebar: User List */}
      <div className="w-1/4 border-r border-gray-800 pr-6">
        <h2 className="text-xl font-bold mb-6 text-blue-400">Target Users</h2>

        <form onSubmit={handleCreateUser} className="mb-8">
          <div className="flex flex-col gap-2">
            <input 
              type="text" 
              placeholder="Enter User ID (e.g. U99)"
              value={newUserId}
              onChange={(e) => setNewUserId(e.target.value.toUpperCase())}
              className="bg-gray-900 border border-gray-700 rounded px-3 py-2 text-sm focus:border-blue-500 outline-none"
            />
            <button type="submit" className="bg-blue-600 hover:bg-blue-500 text-xs font-bold py-2 rounded transition-colors">
              + ADD NEW USER
            </button>
          </div>
        </form>

        <div className="space-y-3">
          {users.map(u => (
            <button 
              key={u.user_id}
              onClick={() => setSelectedUser(u.user_id)}
              className={`w-full text-left p-4 rounded-lg border ${selectedUser === u.user_id ? 'bg-blue-900/30 border-blue-500' : 'bg-gray-900 border-gray-800 hover:border-gray-600'}`}
            >
              <div className="font-mono font-bold">{u.user_id}</div>
              <div className="text-sm text-gray-400">{u.orders} Total Orders</div>
            </button>
          ))}
        </div>
      </div>

      {/* Main Panel: Experiment Simulation */}
      <div className="flex-1 pl-12">
        {!selectedUser ? (
          <div className="h-full flex items-center justify-center text-gray-500 italic">
            Select a user to simulate experiments
          </div>
        ) : (
          <div className="max-w-3xl">
            <header className="flex justify-between items-center mb-10">
              <div>
                <h1 className="text-3xl font-black">User: {selectedUser}</h1>
                <p className="text-gray-400">Current Segments: {exp?.segments.join(', ') || 'None'}</p>
              </div>
              <button 
                onClick={handleSimulateOrders}
                disabled={loading}
                className="bg-orange-600 hover:bg-orange-500 px-6 py-3 rounded-full font-bold transition-all disabled:opacity-50"
              >
                {loading ? 'Processing...' : '🚀 Place 30 Orders'}
              </button>
              <button 
                onClick={handleSync}
                className="bg-blue-600 hover:bg-blue-500 px-6 py-3 rounded-full font-bold ml-2"
              >
                🔄 Sync Segments
              </button>
            </header>

            {/* Dynamic UI Rendering based on Experiment API */}
            <div className="grid gap-6">
              {/* Feature 1: The Pizza Tile */}
              <div className={`p-8 rounded-2xl border-2 transition-all ${exp?.features.show_pizza_tile ? 'border-yellow-500 bg-yellow-500/10' : 'border-gray-800 bg-gray-900 opacity-40'}`}>
                <div className="flex justify-between items-center">
                  <h3 className="text-xl font-bold">🍕 Exclusive Pizza Deal</h3>
                  {!exp?.features.show_pizza_tile && <span className="text-xs uppercase bg-gray-700 px-2 py-1 rounded">Locked</span>}
                </div>
                <p className="mt-2 text-gray-400">This tile only appears for users in the "Power User" segment.</p>
              </div>

              {/* Feature 2: The Banner */}
              <div className="h-32 rounded-2xl bg-gradient-to-r from-blue-600 to-purple-600 flex items-center justify-center relative overflow-hidden">
                <span className="text-2xl font-black italic uppercase tracking-widest">
                  {exp?.features.home_banner || "Standard Welcome"}
                </span>
                {exp?.features.discount_pct && (
                  <div className="absolute top-2 right-2 bg-red-500 text-white text-xs font-bold px-3 py-1 rounded-full animate-bounce">
                    {exp.features.discount_pct}% OFF
                  </div>
                )}
              </div>

              {/* Raw JSON Debugger */}
              <div className="mt-8">
                <h4 className="text-xs font-mono text-gray-500 uppercase mb-2">Raw Experiment Payload</h4>
                <pre className="bg-black p-4 rounded border border-gray-800 text-green-400 text-sm overflow-auto">
                  {JSON.stringify(exp, null, 2)}
                </pre>
              </div>
            </div>
          </div>
        )}
      </div>
    </main>
  );
}