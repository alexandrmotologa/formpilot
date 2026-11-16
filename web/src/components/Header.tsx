import React, { useState } from 'react';
import {
  Activity,
  Cpu,
  Play,
  Database,
  RefreshCw,
  Zap,
  Sliders,
  Globe,
  Download,
  ChevronDown,
} from 'lucide-react';
import { HealthInfo } from '../types';

interface HeaderProps {
  health: HealthInfo | null;
  sseConnected: boolean;
  onOpenSimulator: () => void;
  onOpenAccounts: () => void;
  onOpenSettings: () => void;
  onOpenWebhooks: () => void;
  onExport: (format: 'csv' | 'json') => void;
  onRefresh: () => void;
  isRefreshing: boolean;
}

export const Header: React.FC<HeaderProps> = ({
  health,
  sseConnected,
  onOpenSimulator,
  onOpenAccounts,
  onOpenSettings,
  onOpenWebhooks,
  onExport,
  onRefresh,
  isRefreshing,
}) => {
  const [showExportMenu, setShowExportMenu] = useState(false);

  return (
    <header className="border-b border-slate-800 bg-slate-900/80 backdrop-blur-md sticky top-0 z-30 px-6 py-3.5">
      <div className="max-w-7xl mx-auto flex items-center justify-between">
        {/* Left: Branding */}
        <div className="flex items-center gap-3">
          <div className="w-9 h-9 rounded-lg bg-emerald-500/10 border border-emerald-500/30 flex items-center justify-center text-emerald-400 font-bold shadow-lg shadow-emerald-500/10">
            <Zap className="w-5 h-5 text-emerald-400" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <span className="font-bold tracking-tight text-lg text-white">formpilot</span>
              <span className="text-[11px] font-medium tracking-wide uppercase px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                v1.0 Go Daemon
              </span>
            </div>
            <p className="text-xs text-slate-400">Autonomous Unstructured-to-CRM Pipeline & RevOps Studio</p>
          </div>
        </div>

        {/* Center: Live Architecture Benchmarks */}
        <div className="hidden lg:flex items-center gap-4 bg-slate-950/60 border border-slate-800/80 rounded-full px-4 py-1.5 text-xs text-slate-300">
          <div className="flex items-center gap-1.5">
            <Activity className="w-3.5 h-3.5 text-emerald-400" />
            <span>Target Latency:</span>
            <span className="font-semibold text-emerald-400">&lt; 3.00s</span>
          </div>
          <div className="h-3 w-px bg-slate-800" />
          <div className="flex items-center gap-1.5">
            <Cpu className="w-3.5 h-3.5 text-blue-400" />
            <span>Memory Target:</span>
            <span className="font-semibold text-blue-400">&lt; 15MB RAM</span>
            {health && (
              <span className="text-slate-400 text-[11px]">({health.memory_sys_mb})</span>
            )}
          </div>
          <div className="h-3 w-px bg-slate-800" />
          <div className="flex items-center gap-1.5">
            <span className={`w-2 h-2 rounded-full ${sseConnected ? 'bg-emerald-400 animate-pulse' : 'bg-amber-400'}`} />
            <span className="text-slate-300">{sseConnected ? 'Live SSE' : 'Reconnecting...'}</span>
          </div>
        </div>

        {/* Right: Actions */}
        <div className="flex items-center gap-2">
          {/* Webhook Gateway Button */}
          <button
            onClick={onOpenWebhooks}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-slate-300 hover:text-white bg-slate-800/80 hover:bg-slate-700/80 border border-slate-700/60 rounded-lg transition"
            title="Webhook Gateway & cURL Playground"
          >
            <Globe className="w-3.5 h-3.5 text-blue-400" />
            <span className="hidden sm:inline">Webhooks</span>
          </button>

          {/* Settings / ICP Tuning Button */}
          <button
            onClick={onOpenSettings}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-slate-300 hover:text-white bg-slate-800/80 hover:bg-slate-700/80 border border-slate-700/60 rounded-lg transition"
            title="Configure ICP Scoring Weights"
          >
            <Sliders className="w-3.5 h-3.5 text-emerald-400" />
            <span className="hidden sm:inline">ICP Rules</span>
          </button>

          {/* CRM Cache Button */}
          <button
            onClick={onOpenAccounts}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-slate-300 hover:text-white bg-slate-800/80 hover:bg-slate-700/80 border border-slate-700/60 rounded-lg transition"
            title="View Deduplication Account Cache"
          >
            <Database className="w-3.5 h-3.5 text-purple-400" />
            <span className="hidden sm:inline">CRM Cache</span>
          </button>

          {/* Export Dropdown */}
          <div className="relative">
            <button
              onClick={() => setShowExportMenu(!showExportMenu)}
              className="flex items-center gap-1 px-2.5 py-1.5 text-xs font-medium text-slate-300 hover:text-white bg-slate-800/80 hover:bg-slate-700/80 border border-slate-700/60 rounded-lg transition"
              title="Export Leads"
            >
              <Download className="w-3.5 h-3.5 text-amber-400" />
              <ChevronDown className="w-3 h-3" />
            </button>

            {showExportMenu && (
              <div className="absolute right-0 mt-1 w-32 bg-slate-900 border border-slate-800 rounded-lg shadow-xl py-1 z-40 text-xs">
                <button
                  onClick={() => {
                    onExport('csv');
                    setShowExportMenu(false);
                  }}
                  className="w-full text-left px-3 py-1.5 text-slate-300 hover:bg-slate-800 hover:text-white flex items-center justify-between"
                >
                  <span>Export CSV</span>
                  <span className="text-[10px] text-slate-500 font-mono">.csv</span>
                </button>
                <button
                  onClick={() => {
                    onExport('json');
                    setShowExportMenu(false);
                  }}
                  className="w-full text-left px-3 py-1.5 text-slate-300 hover:bg-slate-800 hover:text-white flex items-center justify-between"
                >
                  <span>Export JSON</span>
                  <span className="text-[10px] text-slate-500 font-mono">.json</span>
                </button>
              </div>
            )}
          </div>

          {/* Refresh Button */}
          <button
            onClick={onRefresh}
            disabled={isRefreshing}
            className="p-1.5 text-slate-400 hover:text-slate-200 bg-slate-800/80 hover:bg-slate-700/80 border border-slate-700/60 rounded-lg transition"
            title="Refresh Leads"
          >
            <RefreshCw className={`w-4 h-4 ${isRefreshing ? 'animate-spin text-emerald-400' : ''}`} />
          </button>

          {/* Simulate Button */}
          <button
            onClick={onOpenSimulator}
            className="flex items-center gap-1.5 px-3.5 py-1.5 text-xs font-semibold text-white bg-emerald-600 hover:bg-emerald-500 rounded-lg shadow-sm shadow-emerald-950 transition"
          >
            <Play className="w-3.5 h-3.5 fill-current" />
            <span>Simulate Lead</span>
          </button>
        </div>
      </div>
    </header>
  );
};
