import React, { useState } from 'react';
import { Search, Flame, Building, ShieldCheck, Clock, Users, DollarSign, ChevronRight } from 'lucide-react';
import { Lead } from '../types';

interface LiveLeadFeedProps {
  leads: Lead[];
  selectedLeadId: string | null;
  onSelectLead: (lead: Lead) => void;
  isLoading: boolean;
}

export const LiveLeadFeed: React.FC<LiveLeadFeedProps> = ({
  leads,
  selectedLeadId,
  onSelectLead,
  isLoading,
}) => {
  const [search, setSearch] = useState('');
  const [filterTier, setFilterTier] = useState<string>('all');

  const filtered = leads.filter((lead) => {
    const query = search.toLowerCase();
    const matchQuery =
      lead.company_name.toLowerCase().includes(query) ||
      lead.domain.toLowerCase().includes(query) ||
      lead.contact_name.toLowerCase().includes(query) ||
      lead.subject.toLowerCase().includes(query);

    if (!matchQuery) return false;
    if (filterTier === 'all') return true;
    return lead.qualification_tier === filterTier;
  });

  const getTierBadge = (tier: string) => {
    switch (tier) {
      case 'Enterprise Hot Lead':
        return (
          <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[11px] font-semibold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
            <Flame className="w-3 h-3 text-emerald-400" />
            Hot Enterprise
          </span>
        );
      case 'Mid-Market Qualified':
        return (
          <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[11px] font-semibold bg-sky-500/10 text-sky-400 border border-sky-500/20">
            <Building className="w-3 h-3 text-sky-400" />
            Mid-Market
          </span>
        );
      case 'SMB / Self-Serve':
        return (
          <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[11px] font-semibold bg-amber-500/10 text-amber-400 border border-amber-500/20">
            SMB / Self-Serve
          </span>
        );
      default:
        return (
          <span className="inline-flex items-center px-2 py-0.5 rounded-full text-[11px] font-medium bg-slate-800 text-slate-400 border border-slate-700">
            Unqualified
          </span>
        );
    }
  };

  const getScoreColor = (score: number) => {
    if (score >= 80) return 'text-emerald-400 bg-emerald-500/10 border-emerald-500/30';
    if (score >= 60) return 'text-sky-400 bg-sky-500/10 border-sky-500/30';
    if (score >= 40) return 'text-amber-400 bg-amber-500/10 border-amber-500/30';
    return 'text-slate-400 bg-slate-800/40 border-slate-700';
  };

  return (
    <div className="flex flex-col h-full bg-slate-900/40 border border-slate-800 rounded-xl overflow-hidden">
      {/* Search and Filters Header */}
      <div className="p-3.5 border-b border-slate-800 bg-slate-900/60 flex flex-col gap-3">
        <div className="relative">
          <Search className="w-4 h-4 absolute left-3 top-2.5 text-slate-400" />
          <input
            type="text"
            placeholder="Search by company, domain, contact, or RFQ notes..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full pl-9 pr-3 py-1.5 bg-slate-950/70 border border-slate-800 rounded-lg text-xs text-slate-200 placeholder-slate-500 focus:outline-none focus:border-emerald-500/50"
          />
        </div>

        {/* Tier filter tabs */}
        <div className="flex items-center gap-1.5 overflow-x-auto text-xs">
          {[
            { id: 'all', label: 'All Leads' },
            { id: 'Enterprise Hot Lead', label: 'Hot Enterprise' },
            { id: 'Mid-Market Qualified', label: 'Mid-Market' },
            { id: 'SMB / Self-Serve', label: 'SMB' },
            { id: 'Unqualified / Low Intent', label: 'Unqualified' },
          ].map((tab) => (
            <button
              key={tab.id}
              onClick={() => setFilterTier(tab.id)}
              className={`px-2.5 py-1 rounded-md text-[11px] font-medium whitespace-nowrap transition ${
                filterTier === tab.id
                  ? 'bg-slate-800 text-white font-semibold shadow-sm'
                  : 'text-slate-400 hover:text-slate-200 hover:bg-slate-800/40'
              }`}
            >
              {tab.label}
            </button>
          ))}
        </div>
      </div>

      {/* Feed List */}
      <div className="flex-1 overflow-y-auto divide-y divide-slate-800/60 p-2 space-y-1">
        {isLoading && leads.length === 0 ? (
          <div className="py-12 text-center text-slate-400 text-xs">
            Loading real-time lead stream...
          </div>
        ) : filtered.length === 0 ? (
          <div className="py-12 text-center text-slate-500 text-xs">
            No leads match current filter criteria.
          </div>
        ) : (
          filtered.map((lead) => {
            const isSelected = lead.id === selectedLeadId;
            return (
              <div
                key={lead.id}
                onClick={() => onSelectLead(lead)}
                className={`p-3 rounded-lg cursor-pointer transition border ${
                  isSelected
                    ? 'bg-slate-800/80 border-emerald-500/40 shadow-sm'
                    : 'bg-slate-900/40 hover:bg-slate-800/50 border-transparent hover:border-slate-800'
                }`}
              >
                <div className="flex items-start justify-between gap-3">
                  {/* Left: Lead Details */}
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-1 flex-wrap">
                      <span className="font-semibold text-sm text-slate-100 truncate">
                        {lead.company_name || 'Unspecified Company'}
                      </span>
                      {lead.domain && (
                        <span className="text-[11px] text-slate-400 font-mono">
                          {lead.domain}
                        </span>
                      )}
                      {lead.dedup_status === 'existing_matched' ? (
                        <span className="inline-flex items-center gap-1 text-[10px] font-semibold text-purple-400 bg-purple-500/10 border border-purple-500/20 px-1.5 py-0.2 rounded">
                          <ShieldCheck className="w-3 h-3 text-purple-400" />
                          Existing Account Matched
                        </span>
                      ) : (
                        <span className="text-[10px] text-emerald-400/80 bg-emerald-500/5 px-1.5 py-0.2 rounded">
                          New Account
                        </span>
                      )}
                    </div>

                    <div className="text-xs text-slate-300 font-medium truncate mb-1">
                      {lead.contact_name ? `${lead.contact_name} ` : ''}
                      <span className="text-slate-400 text-[11px]">
                        {lead.contact_email ? `(${lead.contact_email})` : ''}
                      </span>
                    </div>

                    <p className="text-[11px] text-slate-400 line-clamp-1 italic mb-2">
                      &quot;{lead.subject}&quot;
                    </p>

                    {/* Metadata Pills */}
                    <div className="flex items-center gap-2 text-[11px] text-slate-400 flex-wrap">
                      {lead.requested_seats > 0 && (
                        <span className="inline-flex items-center gap-1 bg-slate-950/60 px-2 py-0.5 rounded border border-slate-800/80">
                          <Users className="w-3 h-3 text-slate-400" />
                          {lead.requested_seats} seats
                        </span>
                      )}

                      {lead.budget > 0 && (
                        <span className="inline-flex items-center gap-1 bg-slate-950/60 px-2 py-0.5 rounded border border-slate-800/80">
                          <DollarSign className="w-3 h-3 text-emerald-400" />
                          ${lead.budget >= 1000 ? `${(lead.budget / 1000).toFixed(0)}k` : lead.budget}
                        </span>
                      )}

                      <span className="inline-flex items-center gap-1 text-slate-400">
                        <Clock className="w-3 h-3 text-slate-400" />
                        {lead.latency_ms}ms
                      </span>

                      <span className="text-slate-400 font-mono text-[10px]">
                        {lead.source}
                      </span>
                    </div>
                  </div>

                  {/* Right: ICP Score & Tier */}
                  <div className="flex flex-col items-end gap-1.5">
                    <div
                      className={`w-10 h-10 rounded-lg flex items-center justify-center font-bold text-sm border ${getScoreColor(
                        lead.icp_score
                      )}`}
                      title={`ICP Qualification Fit Score: ${lead.icp_score}/100`}
                    >
                      {lead.icp_score}
                    </div>
                    {getTierBadge(lead.qualification_tier)}
                    <ChevronRight className="w-4 h-4 text-slate-400 mt-1" />
                  </div>
                </div>
              </div>
            );
          })
        )}
      </div>
    </div>
  );
};
