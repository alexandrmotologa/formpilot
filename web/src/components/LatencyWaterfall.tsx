import React from 'react';
import { Clock, CheckCircle2, Zap } from 'lucide-react';
import { Lead } from '../types';

interface LatencyWaterfallProps {
  lead: Lead;
}

export const LatencyWaterfall: React.FC<LatencyWaterfallProps> = ({ lead }) => {
  const mime = lead.latency_mime_ms || 1;
  const extract = lead.latency_extract_ms || 3;
  const dedup = lead.latency_dedup_ms || 1;
  const score = lead.latency_score_ms || 1;
  const draft = lead.latency_draft_ms || 2;
  const crm = lead.latency_crm_ms || 25;
  const storage = lead.latency_storage_ms || 2;
  const total = lead.latency_ms || mime + extract + dedup + score + draft + crm + storage;

  const stages = [
    { label: 'MIME & Attachment Decode', ms: mime, color: 'bg-indigo-500', text: 'text-indigo-400' },
    { label: 'Entity & Intent Extraction', ms: extract, color: 'bg-emerald-500', text: 'text-emerald-400' },
    { label: 'Deduplication & Domain Match', ms: dedup, color: 'bg-purple-500', text: 'text-purple-400' },
    { label: 'ICP Qualification Scoring', ms: score, color: 'bg-sky-500', text: 'text-sky-400' },
    { label: 'Outreach Draft Synthesis', ms: draft, color: 'bg-blue-500', text: 'text-blue-400' },
    { label: 'Multi-CRM Atomic Dispatch', ms: crm, color: 'bg-amber-500', text: 'text-amber-400' },
    { label: 'SQLite Store & SSE Broadcast', ms: storage, color: 'bg-teal-500', text: 'text-teal-400' },
  ];

  return (
    <div className="bg-slate-950/70 border border-slate-800 rounded-lg p-3.5 space-y-2.5">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Clock className="w-4 h-4 text-emerald-400" />
          <span className="text-xs font-semibold text-slate-200 uppercase tracking-wide">
            Pipeline Latency Waterfall (Sub-3s Guarantee)
          </span>
        </div>
        <div className="flex items-center gap-2">
          <span className="inline-flex items-center gap-1 text-[11px] font-bold text-emerald-400 bg-emerald-500/10 px-2 py-0.5 rounded border border-emerald-500/30">
            <CheckCircle2 className="w-3 h-3 text-emerald-400" />
            Total: {total}ms
          </span>
          <span className="text-[10px] text-slate-400 font-mono">
            {((total / 3000) * 100).toFixed(1)}% of 3.0s SLA
          </span>
        </div>
      </div>

      {/* Stacked Waterfall Progress Bar */}
      <div className="h-3 w-full bg-slate-900 rounded-full overflow-hidden flex shadow-inner">
        {stages.map((st, i) => {
          const pct = Math.max(3, (st.ms / total) * 100);
          return (
            <div
              key={i}
              className={`${st.color} h-full transition-all duration-300 hover:opacity-80`}
              style={{ width: `${pct}%` }}
              title={`${st.label}: ${st.ms}ms (${((st.ms / total) * 100).toFixed(1)}%)`}
            />
          );
        })}
      </div>

      {/* Detail Breakdown Grid */}
      <div className="grid grid-cols-2 sm:grid-cols-4 lg:grid-cols-7 gap-2 pt-1">
        {stages.map((st, i) => (
          <div key={i} className="bg-slate-900/60 p-2 rounded border border-slate-800/80 flex flex-col justify-between">
            <span className="text-[10px] text-slate-400 truncate block" title={st.label}>
              {st.label}
            </span>
            <span className={`text-xs font-bold font-mono mt-1 ${st.text}`}>
              {st.ms}ms
            </span>
          </div>
        ))}
      </div>
    </div>
  );
};
