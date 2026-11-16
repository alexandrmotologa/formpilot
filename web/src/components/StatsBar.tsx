import React from 'react';
import { Flame, ShieldCheck, Clock, CheckCircle2, Building, Inbox } from 'lucide-react';
import { StatsSummary } from '../types';

interface StatsBarProps {
  stats: StatsSummary | null;
}

export const StatsBar: React.FC<StatsBarProps> = ({ stats }) => {
  return (
    <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-3 mb-6">
      {/* Total Leads */}
      <div className="bg-slate-900/60 border border-slate-800 rounded-xl p-3.5 flex flex-col justify-between">
        <div className="flex items-center justify-between text-slate-400 mb-2">
          <span className="text-xs font-medium">Total Inbound</span>
          <Inbox className="w-4 h-4 text-slate-400" />
        </div>
        <div>
          <div className="text-2xl font-bold text-white tracking-tight">
            {stats?.total_leads ?? 0}
          </div>
          <div className="text-[11px] text-slate-400 mt-0.5">Autonomous triage</div>
        </div>
      </div>

      {/* Enterprise Hot Leads */}
      <div className="bg-slate-900/60 border border-emerald-950/60 rounded-xl p-3.5 flex flex-col justify-between relative overflow-hidden">
        <div className="absolute top-0 right-0 w-16 h-16 bg-emerald-500/5 rounded-full blur-xl pointer-events-none" />
        <div className="flex items-center justify-between text-emerald-400 mb-2">
          <span className="text-xs font-medium">Enterprise Hot</span>
          <Flame className="w-4 h-4 text-emerald-400" />
        </div>
        <div>
          <div className="text-2xl font-bold text-emerald-400 tracking-tight">
            {stats?.enterprise_hot_leads ?? 0}
          </div>
          <div className="text-[11px] text-emerald-500/80 mt-0.5">Score &gt;= 80 (P0 AE Alert)</div>
        </div>
      </div>

      {/* Mid-Market Qualified */}
      <div className="bg-slate-900/60 border border-slate-800 rounded-xl p-3.5 flex flex-col justify-between">
        <div className="flex items-center justify-between text-sky-400 mb-2">
          <span className="text-xs font-medium">Mid-Market</span>
          <Building className="w-4 h-4 text-sky-400" />
        </div>
        <div>
          <div className="text-2xl font-bold text-sky-400 tracking-tight">
            {stats?.mid_market_leads ?? 0}
          </div>
          <div className="text-[11px] text-slate-400 mt-0.5">Score 60 - 79</div>
        </div>
      </div>

      {/* Deduplicated Accounts */}
      <div className="bg-slate-900/60 border border-purple-950/60 rounded-xl p-3.5 flex flex-col justify-between">
        <div className="flex items-center justify-between text-purple-400 mb-2">
          <span className="text-xs font-medium">Dedup Resolved</span>
          <ShieldCheck className="w-4 h-4 text-purple-400" />
        </div>
        <div>
          <div className="text-2xl font-bold text-purple-400 tracking-tight">
            {stats?.deduplicated_accounts ?? 0}
          </div>
          <div className="text-[11px] text-purple-400/80 mt-0.5">Prevented duplicate reps</div>
        </div>
      </div>

      {/* CRM Syncs */}
      <div className="bg-slate-900/60 border border-slate-800 rounded-xl p-3.5 flex flex-col justify-between">
        <div className="flex items-center justify-between text-emerald-400 mb-2">
          <span className="text-xs font-medium">CRM Dispatched</span>
          <CheckCircle2 className="w-4 h-4 text-emerald-400" />
        </div>
        <div>
          <div className="text-2xl font-bold text-slate-100 tracking-tight">
            {stats?.successful_syncs ?? 0}
          </div>
          <div className="text-[11px] text-slate-400 mt-0.5">100% Idempotent</div>
        </div>
      </div>

      {/* Ingestion Latency */}
      <div className="bg-slate-900/60 border border-slate-800 rounded-xl p-3.5 flex flex-col justify-between">
        <div className="flex items-center justify-between text-amber-400 mb-2">
          <span className="text-xs font-medium">Avg Latency</span>
          <Clock className="w-4 h-4 text-amber-400" />
        </div>
        <div>
          <div className="text-2xl font-bold text-amber-400 tracking-tight">
            {stats?.average_latency_ms ? `${stats.average_latency_ms.toFixed(0)}ms` : '0ms'}
          </div>
          <div className="text-[11px] text-emerald-400 mt-0.5 font-medium">&lt; 3.0s SLA target</div>
        </div>
      </div>
    </div>
  );
};
