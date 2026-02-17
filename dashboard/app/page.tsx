"use client";
import { useState, useEffect } from 'react';
import { api, ExperimentResponse } from '@/lib/api';

export default function Dashboard() {
  const [users, setUsers] = useState<{ user_id: string, orders: number }[]>([]);
  const [selectedUser, setSelectedUser] = useState<string | null>(null);
  const [exp, setExp] = useState<ExperimentResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [newUserId, setNewUserId] = useState('');
  const [orderCount, setOrderCount] = useState<number>(1);
  const [isInstant, setIsInstant] = useState(true);

  const refreshUsers = async () => {
    const data = await api.getUsers();
    setUsers(data);
  };

  useEffect(() => { refreshUsers(); }, []);

  useEffect(() => {
    if (selectedUser) {
      api.getExperiments(selectedUser).then(setExp);
    }
  }, [selectedUser]);

  const handleSimulateOrders = async () => {
    if (!selectedUser) return;
    setLoading(true);
    await api.placeOrders(selectedUser, orderCount, isInstant);

    if (isInstant) {
      setTimeout(async () => {
        await refreshUsers();
        const newExp = await api.getExperiments(selectedUser);
        setExp(newExp);
        setLoading(false);
      }, 1000);
    } else {
      setLoading(false);
    }
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
    await refreshUsers();
  };

  return (
    <div className="min-h-screen bg-[#080808] text-white font-sans">
      <header className="flex items-center justify-between p-6 bg-[#080808] border-b border-[#2DD283]/10">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 bg-[#2DD283] rounded-xl flex items-center justify-center shadow-[0_0_20px_rgba(45,210,131,0.3)]">
            <span className="text-black font-black text-2xl mt-[-2px]">S</span>
          </div>
          <div>
            <h1 className="text-xl font-black tracking-tight text-white uppercase italic leading-none">
              Swish <span className="text-[#2DD283]">Labs</span>
            </h1>
            <p className="text-[10px] text-[#2DD283] font-bold tracking-widest uppercase opacity-80 mt-1">
              Delight Delivered in 10 Mins
            </p>
          </div>
        </div>

        <div className="flex items-center gap-4 text-right">
          <button 
            onClick={handleSync}
            className="text-[10px] font-bold text-[#2DD283] border border-[#2DD283]/30 px-3 py-1 rounded-full hover:bg-[#2DD283]/10 transition-all"
          >
            🔄 FORCE SYNC ENGINE
          </button>
          <div className="px-3 py-1 bg-[#2DD283]/10 rounded-full border border-[#2DD283]/20">
            <span className="text-[#2DD283] text-[10px] font-bold">SYSTEM ACTIVE</span>
          </div>
        </div>
      </header>

      <main className="flex p-8 gap-8">
        {/* Sidebar: User List */}
        <div className="w-1/4">
          <h2 className="text-[#2DD283] font-black uppercase italic tracking-widest text-sm mb-4">Target Users</h2>

          <form onSubmit={handleCreateUser} className="mb-6">
            <div className="flex flex-col gap-2">
              <input
                type="text"
                placeholder="User ID (e.g. U99)"
                value={newUserId}
                onChange={(e) => setNewUserId(e.target.value.toUpperCase())}
                className="bg-[#121212] border border-white/10 rounded-xl px-4 py-3 text-sm focus:border-[#2DD283] outline-none transition-all" />
              <button type="submit" className="bg-white text-black hover:bg-[#2DD283] transition-colors text-xs font-black py-3 rounded-xl uppercase tracking-tighter">
                + Enroll New User
              </button>
            </div>
          </form>

          <div className="space-y-2 max-h-[60vh] overflow-y-auto pr-2 custom-scrollbar">
            {users.map(u => (
              <button
                key={u.user_id}
                onClick={() => setSelectedUser(u.user_id)}
                className={`w-full text-left p-4 rounded-2xl border transition-all ${selectedUser === u.user_id ? 'bg-[#2DD283]/10 border-[#2DD283] shadow-[0_0_15px_rgba(45,210,131,0.1)]' : 'bg-[#121212] border-white/5 hover:border-white/20'}`}
              >
                <div className="flex justify-between items-center">
                  <span className="font-black italic text-lg">{u.user_id}</span>
                  <span className={`text-[10px] font-bold px-2 py-0.5 rounded ${u.orders >= 25 ? 'bg-[#2DD283] text-black' : 'bg-white/10 text-white/50'}`}>
                    {u.orders >= 25 ? 'POWER' : 'BASIC'}
                  </span>
                </div>
                <div className="text-xs font-medium text-gray-500 mt-1">{u.orders} Total Orders</div>
              </button>
            ))}
          </div>
        </div>

        {/* Main Panel */}
        <div className="flex-1 bg-[#121212] rounded-[32px] p-10 border border-white/5">
          {!selectedUser ? (
            <div className="h-full flex flex-col items-center justify-center text-gray-600">
              <div className="w-16 h-16 border-4 border-dashed border-gray-800 rounded-full mb-4 animate-spin-slow"></div>
              <p className="italic font-medium">Select a Delight Centre User to simulate experiments</p>
            </div>
          ) : (
            <div className="max-w-3xl mx-auto">
              <div className="flex justify-between items-start mb-12">
                <div>
                  <div className="text-[#2DD283] text-[10px] font-black tracking-[0.3em] uppercase mb-2">Active Session</div>
                  <h1 className="text-5xl font-black italic tracking-tighter">USER: {selectedUser}</h1>
                </div>
                <div className="text-right">
                  <div className="text-gray-500 text-[10px] font-bold uppercase mb-1">Assigned Segments</div>
                  <div className="flex gap-2 justify-end">
                    {exp?.segments.length ? exp.segments.map(s => (
                      <span key={s} className="bg-[#2DD283] text-black text-[10px] font-extrabold px-3 py-1 rounded-full">{s.toUpperCase()}</span>
                    )) : <span className="text-gray-700 italic text-sm font-medium">No Active Segments</span>}
                  </div>
                </div>
              </div>

              {/* Order Simulation Control */}
              <div className="bg-black/40 rounded-[24px] p-8 border border-white/5 mb-10">
                <div className="flex items-center gap-6">
                  <div className="flex-1">
                    <label className="block text-[10px] font-black text-gray-500 uppercase tracking-widest mb-3">Load Volume</label>
                    <input
                      type="number"
                      value={orderCount}
                      onChange={(e) => setOrderCount(parseInt(e.target.value) || 0)}
                      className="w-full bg-black border border-white/10 rounded-xl px-5 py-4 text-3xl font-black italic focus:border-[#2DD283] outline-none transition-all text-[#2DD283]" />
                  </div>
                  <div className="flex-1 pt-6">
                    <button
                      onClick={handleSimulateOrders}
                      disabled={loading}
                      className="w-full bg-[#2DD283] hover:bg-[#25b570] text-black px-8 py-5 rounded-2xl font-black italic text-xl transition-all disabled:opacity-50 shadow-[0_10px_30px_rgba(45,210,131,0.2)] active:scale-95"
                    >
                      {loading ? 'SWISHING...' : `🚀 Place ${orderCount} Orders`}
                    </button>
                    <label className="flex items-center justify-center gap-2 mt-4 cursor-pointer group">
                      <input
                        type="checkbox"
                        checked={isInstant}
                        onChange={(e) => setIsInstant(e.target.checked)}
                        className="w-4 h-4 accent-[#2DD283]" />
                      <span className="text-[10px] text-gray-500 font-bold group-hover:text-gray-300 transition-colors uppercase tracking-widest">Enable Hot Path Sync</span>
                    </label>
                  </div>
                </div>
              </div>

              {/* Experiment Features */}
              <div className="grid gap-6">
                <div className={`p-8 rounded-[24px] border-2 transition-all duration-500 ${exp?.features.show_pizza_tile ? 'border-[#2DD283] bg-[#2DD283]/5' : 'border-white/5 bg-white/[0.02] opacity-30 grayscale'}`}>
                  <div className="flex justify-between items-center mb-4">
                    <div className="flex items-center gap-3">
                      <span className="text-3xl">🍕</span>
                      <h3 className="text-xl font-black italic uppercase">Exclusive Pizza Deal</h3>
                    </div>
                    {!exp?.features.show_pizza_tile && <span className="text-[10px] font-black uppercase bg-white/10 px-3 py-1 rounded-full">Locked</span>}
                  </div>
                  <p className="text-sm text-gray-500 font-medium">Power users receive 15% extra delight on all pizza categories.</p>
                </div>

                <div className={`h-40 rounded-[24px] relative overflow-hidden flex items-center justify-center transition-all duration-700 ${exp?.features.home_banner ? 'bg-gradient-to-br from-[#2DD283] to-[#1a8a54]' : 'bg-[#1a1a1a] border border-white/5'}`}>
                  <span className={`text-3xl font-black italic uppercase tracking-tighter ${exp?.features.home_banner ? 'text-black' : 'text-white/10'}`}>
                    {exp?.features.home_banner || "Standard Welcome"}
                  </span>
                  {exp?.features.discount_pct && (
                    <div className="absolute bottom-4 right-6 bg-black text-[#2DD283] text-lg font-black italic px-4 py-1 rounded-lg transform -rotate-3 animate-pulse">
                      {exp.features.discount_pct}% OFF
                    </div>
                  )}
                </div>

                <div className="mt-8">
                  <div className="flex items-center gap-2 mb-3">
                    <div className="w-1.5 h-1.5 bg-[#2DD283] rounded-full"></div>
                    <h4 className="text-[10px] font-black text-gray-500 uppercase tracking-widest">Engine Response Debugger</h4>
                  </div>
                  <pre className="bg-black/60 p-6 rounded-2xl border border-white/5 text-[#2DD283] text-xs font-mono overflow-auto leading-relaxed shadow-inner">
                    {JSON.stringify(exp, null, 2)}
                  </pre>
                </div>
              </div>
            </div>
          )}
        </div>
      </main>
    </div>
  );
}