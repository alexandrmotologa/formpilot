import React, { useState, useEffect } from 'react';
import { X, Sliders, Save, RefreshCw, Check, Sparkles } from 'lucide-react';
import { ICPConfig } from '../types';

interface SettingsModalProps {
  isOpen: boolean;
  onClose: () => void;
  onRefreshStats: () => void;
}

export const SettingsModal: React.FC<SettingsModalProps> = ({
  isOpen,
  onClose,
  onRefreshStats,
}) => {
  const [config, setConfig] = useState<ICPConfig>({
    min_qualification_score: 60,
    weights: {
      corporate_domain: 20,
      seat_volume: 25,
      seniority: 20,
      urgency: 15,
      budget: 10,
      product_fit: 10,
    },
  });
  const [isSaving, setIsSaving] = useState(false);
  const [isRecalculating, setIsRecalculating] = useState(false);
  const [savedSuccess, setSavedSuccess] = useState(false);
  const [recalcSummary, setRecalcSummary] = useState<string | null>(null);

  useEffect(() => {
    if (isOpen) {
      fetch('/api/v1/config/icp')
        .then((res) => res.json())
        .then((data) => {
          if (data && data.weights) {
            setConfig(data);
          }
        })
        .catch(() => {});
    }
  }, [isOpen]);

  if (!isOpen) return null;

  const handleWeightChange = (key: keyof ICPConfig['weights'], value: number) => {
    setConfig((prev) => ({
      ...prev,
      weights: {
        ...prev.weights,
        [key]: value,
      },
    }));
  };

  const handleSave = async () => {
    setIsSaving(true);
    setSavedSuccess(false);
    try {
      const res = await fetch('/api/v1/config/icp', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(config),
      });
      if (res.ok) {
        setSavedSuccess(true);
        setTimeout(() => setSavedSuccess(false), 2500);
      }
    } finally {
      setIsSaving(false);
    }
  };

  const handleRecalculateAll = async () => {
    setIsRecalculating(true);
    setRecalcSummary(null);
    try {
      // First save current weights
      await fetch('/api/v1/config/icp', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(config),
      });

      const res = await fetch('/api/v1/leads/recalculate-all', {
        method: 'POST',
      });
      if (res.ok) {
        const data = await res.json();
        setRecalcSummary(`Successfully re-evaluated ${data.leads_evaluated} leads across current weights.`);
        onRefreshStats();
      }
    } finally {
      setIsRecalculating(false);
    }
  };

  const totalPoints =
    config.weights.corporate_domain +
    config.weights.seat_volume +
    config.weights.seniority +
    config.weights.urgency +
    config.weights.budget +
    config.weights.product_fit;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/80 backdrop-blur-sm animate-in fade-in duration-150">
      <div className="bg-slate-900 border border-slate-800 rounded-xl w-full max-w-xl max-h-[90vh] flex flex-col shadow-2xl overflow-hidden">
        {/* Header */}
        <div className="p-4 border-b border-slate-800 flex items-center justify-between bg-slate-900/90">
          <div className="flex items-center gap-2">
            <div className="w-7 h-7 rounded-lg bg-emerald-500/10 border border-emerald-500/30 flex items-center justify-center text-emerald-400">
              <Sliders className="w-4 h-4" />
            </div>
            <div>
              <h3 className="font-bold text-sm text-white">ICP Scoring Rules & Weights</h3>
              <p className="text-xs text-slate-400">Live dynamic tuning of lead qualification thresholds</p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="p-1 text-slate-400 hover:text-slate-200 bg-slate-800 hover:bg-slate-700 rounded-md transition"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-5 space-y-4">
          <div className="flex items-center justify-between bg-slate-950 p-3 rounded-lg border border-slate-800">
            <div>
              <span className="text-xs font-semibold text-white block">Maximum Qualification Cap</span>
              <span className="text-[11px] text-slate-400">Sum of configured dimension weights</span>
            </div>
            <span
              className={`text-sm font-bold font-mono px-2.5 py-1 rounded ${
                totalPoints === 100
                  ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/30'
                  : 'bg-amber-500/10 text-amber-400 border border-amber-500/30'
              }`}
            >
              {totalPoints} / 100 pts
            </span>
          </div>

          {/* Sliders */}
          <div className="space-y-3.5">
            {/* Corporate Domain */}
            <div>
              <div className="flex justify-between text-xs mb-1">
                <span className="text-slate-300 font-medium">Corporate Domain Weight</span>
                <span className="font-mono text-emerald-400">{config.weights.corporate_domain} pts</span>
              </div>
              <input
                type="range"
                min="0"
                max="40"
                value={config.weights.corporate_domain}
                onChange={(e) => handleWeightChange('corporate_domain', parseInt(e.target.value))}
                className="w-full accent-emerald-500"
              />
            </div>

            {/* Seat Volume */}
            <div>
              <div className="flex justify-between text-xs mb-1">
                <span className="text-slate-300 font-medium">Seat Volume Weight</span>
                <span className="font-mono text-emerald-400">{config.weights.seat_volume} pts</span>
              </div>
              <input
                type="range"
                min="0"
                max="50"
                value={config.weights.seat_volume}
                onChange={(e) => handleWeightChange('seat_volume', parseInt(e.target.value))}
                className="w-full accent-emerald-500"
              />
            </div>

            {/* Seniority */}
            <div>
              <div className="flex justify-between text-xs mb-1">
                <span className="text-slate-300 font-medium">Decision Maker Seniority Weight</span>
                <span className="font-mono text-emerald-400">{config.weights.seniority} pts</span>
              </div>
              <input
                type="range"
                min="0"
                max="40"
                value={config.weights.seniority}
                onChange={(e) => handleWeightChange('seniority', parseInt(e.target.value))}
                className="w-full accent-emerald-500"
              />
            </div>

            {/* Urgency */}
            <div>
              <div className="flex justify-between text-xs mb-1">
                <span className="text-slate-300 font-medium">Timeline Urgency Weight</span>
                <span className="font-mono text-emerald-400">{config.weights.urgency} pts</span>
              </div>
              <input
                type="range"
                min="0"
                max="30"
                value={config.weights.urgency}
                onChange={(e) => handleWeightChange('urgency', parseInt(e.target.value))}
                className="w-full accent-emerald-500"
              />
            </div>

            {/* Budget */}
            <div>
              <div className="flex justify-between text-xs mb-1">
                <span className="text-slate-300 font-medium">Budget Viability Weight</span>
                <span className="font-mono text-emerald-400">{config.weights.budget} pts</span>
              </div>
              <input
                type="range"
                min="0"
                max="25"
                value={config.weights.budget}
                onChange={(e) => handleWeightChange('budget', parseInt(e.target.value))}
                className="w-full accent-emerald-500"
              />
            </div>

            {/* Product Fit */}
            <div>
              <div className="flex justify-between text-xs mb-1">
                <span className="text-slate-300 font-medium">Product & Compliance Alignment</span>
                <span className="font-mono text-emerald-400">{config.weights.product_fit} pts</span>
              </div>
              <input
                type="range"
                min="0"
                max="25"
                value={config.weights.product_fit}
                onChange={(e) => handleWeightChange('product_fit', parseInt(e.target.value))}
                className="w-full accent-emerald-500"
              />
            </div>
          </div>

          {/* Feedback */}
          {savedSuccess && (
            <div className="bg-emerald-500/10 border border-emerald-500/30 text-emerald-400 p-2.5 rounded-lg text-xs flex items-center gap-2">
              <Check className="w-4 h-4 shrink-0" />
              <span>Configuration saved and active in pipeline.</span>
            </div>
          )}

          {recalcSummary && (
            <div className="bg-sky-500/10 border border-sky-500/30 text-sky-400 p-2.5 rounded-lg text-xs flex items-center gap-2">
              <Sparkles className="w-4 h-4 shrink-0" />
              <span>{recalcSummary}</span>
            </div>
          )}
        </div>

        {/* Footer Actions */}
        <div className="p-4 border-t border-slate-800 bg-slate-900/90 flex items-center justify-between gap-3">
          <button
            onClick={handleRecalculateAll}
            disabled={isRecalculating}
            className="px-3.5 py-1.5 bg-slate-800 hover:bg-slate-700 text-slate-200 rounded-lg text-xs font-semibold flex items-center gap-1.5 border border-slate-700 transition"
          >
            <RefreshCw className={`w-3.5 h-3.5 ${isRecalculating ? 'animate-spin text-emerald-400' : ''}`} />
            <span>{isRecalculating ? 'Recalculating...' : 'Recalculate All Leads'}</span>
          </button>

          <button
            onClick={handleSave}
            disabled={isSaving}
            className="px-4 py-1.5 bg-emerald-600 hover:bg-emerald-500 text-white rounded-lg text-xs font-semibold flex items-center gap-1.5 transition"
          >
            <Save className="w-3.5 h-3.5" />
            <span>{isSaving ? 'Saving...' : 'Save Weights'}</span>
          </button>
        </div>
      </div>
    </div>
  );
};
