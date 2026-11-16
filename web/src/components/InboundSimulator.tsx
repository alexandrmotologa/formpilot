import React, { useState } from 'react';
import { X, Play, Send, Sparkles, Building, Flame, ShieldAlert, Check } from 'lucide-react';
import { Lead } from '../types';

interface InboundSimulatorProps {
  isOpen: boolean;
  onClose: () => void;
  onSimulate: (preset: string) => Promise<Lead>;
  onCustomSubmit: (sender: string, subject: string, body: string) => Promise<Lead>;
}

export const InboundSimulator: React.FC<InboundSimulatorProps> = ({
  isOpen,
  onClose,
  onSimulate,
  onCustomSubmit,
}) => {
  const [activeTab, setActiveTab] = useState<'presets' | 'custom'>('presets');
  const [sender, setSender] = useState('Marcus Vance <m.vance@apexclearing.com>');
  const [subject, setSubject] = useState('Enterprise RFP: 250 FinTech Seats');
  const [body, setBody] = useState(`Hi Team,
I am Marcus Vance, VP of Engineering at Apex Clearing (apexclearing.com).
We are looking for an automated data pipeline for 250 enterprise seats.
Compliance: SOC2 Type II, SAML SSO with Okta.
Budget is $120,000. Looking to deploy within 30 days.`);
  const [loadingPreset, setLoadingPreset] = useState<string | null>(null);
  const [isSubmittingCustom, setIsSubmittingCustom] = useState(false);
  const [lastResult, setLastResult] = useState<Lead | null>(null);

  if (!isOpen) return null;

  const handleRunPreset = async (preset: string) => {
    setLoadingPreset(preset);
    try {
      const res = await onSimulate(preset);
      setLastResult(res);
    } finally {
      setLoadingPreset(null);
    }
  };

  const handleRunCustom = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmittingCustom(true);
    try {
      const res = await onCustomSubmit(sender, subject, body);
      setLastResult(res);
    } finally {
      setIsSubmittingCustom(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/80 backdrop-blur-sm animate-in fade-in duration-150">
      <div className="bg-slate-900 border border-slate-800 rounded-xl w-full max-w-2xl max-h-[90vh] flex flex-col shadow-2xl overflow-hidden">
        {/* Header */}
        <div className="p-4 border-b border-slate-800 flex items-center justify-between bg-slate-900/90">
          <div className="flex items-center gap-2">
            <div className="w-7 h-7 rounded-lg bg-emerald-500/10 border border-emerald-500/30 flex items-center justify-center text-emerald-400">
              <Play className="w-4 h-4 fill-current" />
            </div>
            <div>
              <h3 className="font-bold text-sm text-white">Inbound Lead & RFQ Simulator</h3>
              <p className="text-xs text-slate-400">Inject real-world sales leads directly into the &lt; 3s Go pipeline</p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="p-1 text-slate-400 hover:text-slate-200 bg-slate-800 hover:bg-slate-700 rounded-md transition"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        {/* Tab switch */}
        <div className="flex border-b border-slate-800 bg-slate-950/50 px-4 text-xs font-medium">
          <button
            onClick={() => setActiveTab('presets')}
            className={`py-2.5 px-3 border-b-2 transition ${
              activeTab === 'presets'
                ? 'border-emerald-500 text-emerald-400 font-semibold'
                : 'border-transparent text-slate-400 hover:text-slate-200'
            }`}
          >
            Realistic Enterprise Presets
          </button>
          <button
            onClick={() => setActiveTab('custom')}
            className={`py-2.5 px-3 border-b-2 transition ${
              activeTab === 'custom'
                ? 'border-emerald-500 text-emerald-400 font-semibold'
                : 'border-transparent text-slate-400 hover:text-slate-200'
            }`}
          >
            Custom RFQ Payload
          </button>
        </div>

        {/* Body */}
        <div className="flex-1 overflow-y-auto p-5 space-y-4">
          {activeTab === 'presets' ? (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
              {/* Preset 1: FinTech Enterprise */}
              <div className="bg-slate-950/80 border border-slate-800 rounded-lg p-3.5 flex flex-col justify-between hover:border-emerald-500/40 transition">
                <div>
                  <div className="flex items-center justify-between mb-1.5">
                    <span className="font-semibold text-xs text-emerald-400 flex items-center gap-1">
                      <Flame className="w-3.5 h-3.5 text-emerald-400" /> FinTech Enterprise RFP
                    </span>
                    <span className="text-[10px] bg-emerald-500/10 text-emerald-400 px-1.5 py-0.5 rounded font-mono">
                      Target: 100/100
                    </span>
                  </div>
                  <p className="text-[11px] text-slate-300 font-medium">Apex Clearing (VP of Engineering)</p>
                  <p className="text-[11px] text-slate-400 mt-1 line-clamp-2">
                    250 seats, $120k budget, SOC2, Okta SAML, immediate 30-day timeline.
                  </p>
                </div>
                <button
                  onClick={() => handleRunPreset('fintech')}
                  disabled={loadingPreset === 'fintech'}
                  className="mt-3 w-full py-1.5 bg-emerald-600/90 hover:bg-emerald-500 text-white rounded text-xs font-semibold flex items-center justify-center gap-1.5 transition"
                >
                  <Play className="w-3 h-3 fill-current" />
                  <span>{loadingPreset === 'fintech' ? 'Injecting...' : 'Dispatch FinTech Lead'}</span>
                </button>
              </div>

              {/* Preset 2: SaaS Mid-Market */}
              <div className="bg-slate-950/80 border border-slate-800 rounded-lg p-3.5 flex flex-col justify-between hover:border-sky-500/40 transition">
                <div>
                  <div className="flex items-center justify-between mb-1.5">
                    <span className="font-semibold text-xs text-sky-400 flex items-center gap-1">
                      <Building className="w-3.5 h-3.5 text-sky-400" /> Mid-Market SaaS Demo
                    </span>
                    <span className="text-[10px] bg-sky-500/10 text-sky-400 px-1.5 py-0.5 rounded font-mono">
                      Target: ~75/100
                    </span>
                  </div>
                  <p className="text-[11px] text-slate-300 font-medium">CloudScale Analytics (Head of Ops)</p>
                  <p className="text-[11px] text-slate-400 mt-1 line-clamp-2">
                    45 seats, $35k budget, HubSpot integration, 2-month timeline.
                  </p>
                </div>
                <button
                  onClick={() => handleRunPreset('saas')}
                  disabled={loadingPreset === 'saas'}
                  className="mt-3 w-full py-1.5 bg-sky-600/90 hover:bg-sky-500 text-white rounded text-xs font-semibold flex items-center justify-center gap-1.5 transition"
                >
                  <Play className="w-3 h-3 fill-current" />
                  <span>{loadingPreset === 'saas' ? 'Injecting...' : 'Dispatch SaaS Demo Lead'}</span>
                </button>
              </div>

              {/* Preset 3: Healthcare HIPAA */}
              <div className="bg-slate-950/80 border border-slate-800 rounded-lg p-3.5 flex flex-col justify-between hover:border-purple-500/40 transition">
                <div>
                  <div className="flex items-center justify-between mb-1.5">
                    <span className="font-semibold text-xs text-purple-400 flex items-center gap-1">
                      <Sparkles className="w-3.5 h-3.5 text-purple-400" /> Healthcare HIPAA RFP
                    </span>
                    <span className="text-[10px] bg-purple-500/10 text-purple-400 px-1.5 py-0.5 rounded font-mono">
                      Target: 100/100
                    </span>
                  </div>
                  <p className="text-[11px] text-slate-300 font-medium">St Jude Health (Chief Medical Officer)</p>
                  <p className="text-[11px] text-slate-400 mt-1 line-clamp-2">
                    500 coordinators, $300k budget, strict BAA & Salesforce Health Cloud.
                  </p>
                </div>
                <button
                  onClick={() => handleRunPreset('hipaa')}
                  disabled={loadingPreset === 'hipaa'}
                  className="mt-3 w-full py-1.5 bg-purple-600/90 hover:bg-purple-500 text-white rounded text-xs font-semibold flex items-center justify-center gap-1.5 transition"
                >
                  <Play className="w-3 h-3 fill-current" />
                  <span>{loadingPreset === 'hipaa' ? 'Injecting...' : 'Dispatch HIPAA Lead'}</span>
                </button>
              </div>

              {/* Preset 4: Disposable Spam */}
              <div className="bg-slate-950/80 border border-slate-800 rounded-lg p-3.5 flex flex-col justify-between hover:border-rose-500/40 transition">
                <div>
                  <div className="flex items-center justify-between mb-1.5">
                    <span className="font-semibold text-xs text-rose-400 flex items-center gap-1">
                      <ShieldAlert className="w-3.5 h-3.5 text-rose-400" /> Disposable Webmail Spam
                    </span>
                    <span className="text-[10px] bg-rose-500/10 text-rose-400 px-1.5 py-0.5 rounded font-mono">
                      Target: 0/100
                    </span>
                  </div>
                  <p className="text-[11px] text-slate-300 font-medium">tempmail.com Backlinks Blast</p>
                  <p className="text-[11px] text-slate-400 mt-1 line-clamp-2">
                    Tests automatic disposable domain penalty (-50) and low-intent routing.
                  </p>
                </div>
                <button
                  onClick={() => handleRunPreset('spam')}
                  disabled={loadingPreset === 'spam'}
                  className="mt-3 w-full py-1.5 bg-rose-900/60 hover:bg-rose-800 text-white rounded text-xs font-semibold flex items-center justify-center gap-1.5 transition"
                >
                  <Play className="w-3 h-3 fill-current" />
                  <span>{loadingPreset === 'spam' ? 'Injecting...' : 'Dispatch Spam Lead'}</span>
                </button>
              </div>
            </div>
          ) : (
            <form onSubmit={handleRunCustom} className="space-y-3">
              <div>
                <label className="block text-xs font-medium text-slate-300 mb-1">From / Sender</label>
                <input
                  type="text"
                  value={sender}
                  onChange={(e) => setSender(e.target.value)}
                  className="w-full px-3 py-1.5 bg-slate-950 border border-slate-800 rounded-lg text-xs text-slate-200 focus:outline-none focus:border-emerald-500"
                  required
                />
              </div>

              <div>
                <label className="block text-xs font-medium text-slate-300 mb-1">Subject</label>
                <input
                  type="text"
                  value={subject}
                  onChange={(e) => setSubject(e.target.value)}
                  className="w-full px-3 py-1.5 bg-slate-950 border border-slate-800 rounded-lg text-xs text-slate-200 focus:outline-none focus:border-emerald-500"
                  required
                />
              </div>

              <div>
                <label className="block text-xs font-medium text-slate-300 mb-1">Body Text / RFQ Notes</label>
                <textarea
                  value={body}
                  onChange={(e) => setBody(e.target.value)}
                  rows={6}
                  className="w-full p-3 bg-slate-950 border border-slate-800 rounded-lg text-xs text-slate-200 focus:outline-none focus:border-emerald-500 font-sans"
                  required
                />
              </div>

              <button
                type="submit"
                disabled={isSubmittingCustom}
                className="w-full py-2 bg-emerald-600 hover:bg-emerald-500 text-white rounded-lg text-xs font-semibold flex items-center justify-center gap-1.5 transition"
              >
                <Send className="w-3.5 h-3.5" />
                <span>{isSubmittingCustom ? 'Processing Pipeline...' : 'Process Custom Lead'}</span>
              </button>
            </form>
          )}

          {/* Last Simulation Notification */}
          {lastResult && (
            <div className="bg-emerald-950/40 border border-emerald-900/60 rounded-lg p-3 text-xs text-emerald-300 flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Check className="w-4 h-4 text-emerald-400 shrink-0" />
                <span>
                  Processed <strong>{lastResult.company_name || 'Inbound'}</strong> in{' '}
                  <span className="font-bold">{lastResult.latency_ms}ms</span> • Score:{' '}
                  <strong>{lastResult.icp_score}/100</strong> ({lastResult.qualification_tier})
                </span>
              </div>
              <span className="text-[11px] text-emerald-400 font-mono">
                {lastResult.dedup_status === 'existing_matched' ? 'Dedup Matched' : 'New Account'}
              </span>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};
