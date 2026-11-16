import React, { useState } from 'react';
import {
  X,
  Copy,
  Check,
  Send,
  Building,
  Mail,
  Phone,
  ShieldAlert,
  ShieldCheck,
  ExternalLink,
  Flame,
  Clock,
  Sparkles,
  Layers,
  FileText,
  RotateCw,
} from 'lucide-react';
import { Lead, ExtractedEntitiesData, SyncLog, AuditEvent } from '../types';

interface LeadDetailDrawerProps {
  lead: Lead | null;
  syncLogs: SyncLog[];
  auditEvents: AuditEvent[];
  onClose: () => void;
  onUpdateDraft: (leadId: string, newDraft: string) => Promise<void>;
  onResync: (leadId: string) => Promise<void>;
}

export const LeadDetailDrawer: React.FC<LeadDetailDrawerProps> = ({
  lead,
  syncLogs,
  auditEvents,
  onClose,
  onUpdateDraft,
  onResync,
}) => {
  const [activeTab, setActiveTab] = useState<'entities' | 'draft' | 'crm' | 'raw' | 'audit'>('entities');
  const [draftText, setDraftText] = useState(lead?.outreach_draft || '');
  const [isCopied, setIsCopied] = useState(false);
  const [isSavingDraft, setIsSavingDraft] = useState(false);
  const [isResyncing, setIsResyncing] = useState(false);

  // Sync state if lead changes
  React.useEffect(() => {
    if (lead) {
      setDraftText(lead.outreach_draft);
    }
  }, [lead]);

  if (!lead) {
    return (
      <div className="h-full bg-slate-900/40 border border-slate-800 rounded-xl p-8 flex flex-col items-center justify-center text-center text-slate-500">
        <Layers className="w-10 h-10 text-slate-700 mb-3" />
        <h3 className="font-semibold text-slate-400 text-sm">Select an inbound lead</h3>
        <p className="text-xs text-slate-500 max-w-xs mt-1">
          Click any lead in the real-time feed to review extracted B2B entities, deduplication match details, and the sales outreach draft.
        </p>
      </div>
    );
  }

  let extractedData: ExtractedEntitiesData | null = null;
  try {
    if (lead.extracted_entities) {
      extractedData = JSON.parse(lead.extracted_entities);
    }
  } catch {
    // fallback
  }

  const handleCopy = () => {
    navigator.clipboard.writeText(draftText);
    setIsCopied(true);
    setTimeout(() => setIsCopied(false), 2000);
  };

  const handleSaveDraft = async () => {
    setIsSavingDraft(true);
    try {
      await onUpdateDraft(lead.id, draftText);
    } finally {
      setIsSavingDraft(false);
    }
  };

  const handleTriggerResync = async () => {
    setIsResyncing(true);
    try {
      await onResync(lead.id);
    } finally {
      setIsResyncing(false);
    }
  };

  return (
    <div className="flex flex-col h-full bg-slate-900/80 border border-slate-800 rounded-xl overflow-hidden backdrop-blur-md">
      {/* Header */}
      <div className="p-4 border-b border-slate-800 bg-slate-900/90 flex items-start justify-between gap-4">
        <div>
          <div className="flex items-center gap-2.5 flex-wrap">
            <h2 className="text-lg font-bold text-white tracking-tight">
              {lead.company_name || 'Inbound Lead'}
            </h2>
            {lead.domain && (
              <a
                href={`https://${lead.domain}`}
                target="_blank"
                rel="noreferrer"
                className="inline-flex items-center gap-1 text-xs text-emerald-400 hover:text-emerald-300 font-mono bg-emerald-500/10 px-2 py-0.5 rounded border border-emerald-500/20"
              >
                {lead.domain}
                <ExternalLink className="w-3 h-3" />
              </a>
            )}
            <span
              className={`text-xs px-2.5 py-0.5 rounded-full font-semibold ${
                lead.icp_score >= 80
                  ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/30'
                  : lead.icp_score >= 60
                  ? 'bg-sky-500/10 text-sky-400 border border-sky-500/30'
                  : 'bg-slate-800 text-slate-400 border border-slate-700'
              }`}
            >
              Score: {lead.icp_score}/100 • {lead.qualification_tier}
            </span>
          </div>

          <div className="flex items-center gap-3 text-xs text-slate-400 mt-1.5 flex-wrap">
            <span className="text-slate-300 font-medium">{lead.contact_name}</span>
            {extractedData?.job_title && (
              <span className="text-slate-400">({extractedData.job_title})</span>
            )}
            <span>•</span>
            <span className="font-mono">{lead.contact_email}</span>
            <span>•</span>
            <span className="inline-flex items-center gap-1 text-emerald-400 font-medium">
              <Clock className="w-3 h-3" /> {lead.latency_ms}ms execution
            </span>
          </div>
        </div>

        <button
          onClick={onClose}
          className="p-1 text-slate-400 hover:text-slate-200 bg-slate-800/80 hover:bg-slate-700 rounded-md transition"
        >
          <X className="w-5 h-5" />
        </button>
      </div>

      {/* Deduplication Status Alert Banner */}
      <div
        className={`px-4 py-2.5 text-xs flex items-center justify-between border-b ${
          lead.dedup_status === 'existing_matched'
            ? 'bg-purple-950/40 border-purple-900/60 text-purple-300'
            : 'bg-emerald-950/30 border-emerald-900/50 text-emerald-300'
        }`}
      >
        <div className="flex items-center gap-2">
          {lead.dedup_status === 'existing_matched' ? (
            <>
              <ShieldCheck className="w-4 h-4 text-purple-400 shrink-0" />
              <span>
                <strong>CRM Deduplication Match:</strong> Associated with existing account{' '}
                <strong>&quot;{lead.matched_account_name}&quot;</strong> (Owner:{' '}
                {lead.matched_account_owner || 'Enterprise AE'}). Prevented duplicate account creation!
              </span>
            </>
          ) : (
            <>
              <Building className="w-4 h-4 text-emerald-400 shrink-0" />
              <span>
                <strong>New Account Provisioned:</strong> No duplicate account found for domain{' '}
                <code>{lead.domain}</code>. Provisioned new corporate account.
              </span>
            </>
          )}
        </div>
        {lead.matched_account_id && (
          <span className="font-mono text-[11px] bg-purple-900/40 px-2 py-0.5 rounded border border-purple-800/60">
            ID: {lead.matched_account_id}
          </span>
        )}
      </div>

      {/* Tabs */}
      <div className="flex items-center border-b border-slate-800 bg-slate-900/50 px-4 text-xs">
        {[
          { id: 'entities', label: 'B2B Entities & ICP' },
          { id: 'draft', label: 'Sales Outreach Draft' },
          { id: 'crm', label: `CRM Dispatch (${syncLogs.length})` },
          { id: 'raw', label: 'Raw Inbound MIME' },
          { id: 'audit', label: `Audit Trail (${auditEvents.length})` },
        ].map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id as typeof activeTab)}
            className={`py-2.5 px-3 border-b-2 font-medium transition whitespace-nowrap ${
              activeTab === tab.id
                ? 'border-emerald-500 text-emerald-400'
                : 'border-transparent text-slate-400 hover:text-slate-200'
            }`}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {/* Tab Content */}
      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        {/* TAB 1: ENTITIES & ICP */}
        {activeTab === 'entities' && (
          <div className="space-y-4">
            {/* Extracted Entity Grid */}
            <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
              <div className="bg-slate-950/60 border border-slate-800/80 rounded-lg p-3">
                <span className="text-[11px] text-slate-400 block mb-1">Volume / Seats</span>
                <span className="text-base font-bold text-white">
                  {lead.requested_seats > 0 ? `${lead.requested_seats} seats` : 'Unspecified'}
                </span>
              </div>

              <div className="bg-slate-950/60 border border-slate-800/80 rounded-lg p-3">
                <span className="text-[11px] text-slate-400 block mb-1">Allocated Budget</span>
                <span className="text-base font-bold text-emerald-400">
                  {lead.budget > 0 ? `$${lead.budget.toLocaleString()}` : 'Unspecified'}
                </span>
              </div>

              <div className="bg-slate-950/60 border border-slate-800/80 rounded-lg p-3">
                <span className="text-[11px] text-slate-400 block mb-1">Timeline Urgency</span>
                <span className="text-base font-bold text-amber-400">
                  {lead.urgency || 'Standard'}
                </span>
              </div>

              <div className="bg-slate-950/60 border border-slate-800/80 rounded-lg p-3">
                <span className="text-[11px] text-slate-400 block mb-1">Seniority Level</span>
                <span className="text-base font-bold text-sky-400">
                  {extractedData?.seniority || 'Unspecified'}
                </span>
              </div>
            </div>

            {/* Products & Tech Stack */}
            <div className="bg-slate-950/60 border border-slate-800/80 rounded-lg p-4 space-y-3">
              <h4 className="text-xs font-semibold text-slate-300 uppercase tracking-wide">
                Identified Products & Enterprise Tech Stack
              </h4>

              <div>
                <span className="text-[11px] text-slate-400 block mb-1.5">Products Requested:</span>
                <div className="flex flex-wrap gap-1.5">
                  {extractedData?.requested_products && extractedData.requested_products.length > 0 ? (
                    extractedData.requested_products.map((p, i) => (
                      <span
                        key={i}
                        className="text-xs bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 px-2 py-0.5 rounded"
                      >
                        {p}
                      </span>
                    ))
                  ) : (
                    <span className="text-xs text-slate-400">Standard platform inquiry</span>
                  )}
                </div>
              </div>

              <div>
                <span className="text-[11px] text-slate-400 block mb-1.5">
                  Security, Compliance & Integrations:
                </span>
                <div className="flex flex-wrap gap-1.5">
                  {extractedData?.tech_stack && extractedData.tech_stack.length > 0 ? (
                    extractedData.tech_stack.map((t, i) => (
                      <span
                        key={i}
                        className="text-xs bg-blue-500/10 text-blue-400 border border-blue-500/20 px-2 py-0.5 rounded"
                      >
                        {t}
                      </span>
                    ))
                  ) : (
                    <span className="text-xs text-slate-400">No specific compliance tags</span>
                  )}
                </div>
              </div>

              {extractedData?.pain_points && (
                <div>
                  <span className="text-[11px] text-slate-400 block mb-1">Extracted Requirements Notes:</span>
                  <p className="text-xs text-slate-300 bg-slate-900 p-2.5 rounded border border-slate-800 font-sans">
                    {extractedData.pain_points}
                  </p>
                </div>
              )}
            </div>

            {/* Contact details */}
            <div className="bg-slate-950/60 border border-slate-800/80 rounded-lg p-4 space-y-2">
              <h4 className="text-xs font-semibold text-slate-300 uppercase tracking-wide">
                Verified Decision Maker
              </h4>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-2 text-xs">
                <div className="flex items-center gap-2 text-slate-300">
                  <Mail className="w-4 h-4 text-slate-400" />
                  <span>{lead.contact_email || 'No email extracted'}</span>
                </div>
                <div className="flex items-center gap-2 text-slate-300">
                  <Phone className="w-4 h-4 text-slate-400" />
                  <span>{lead.contact_phone || 'No direct phone'}</span>
                </div>
                <div className="flex items-center gap-2 text-slate-300">
                  <Building className="w-4 h-4 text-slate-400" />
                  <span>{lead.domain}</span>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* TAB 2: OUTREACH DRAFT STUDIO */}
        {activeTab === 'draft' && (
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <div>
                <h4 className="text-xs font-semibold text-white">Personalized Sales Representative Reply Draft</h4>
                <p className="text-[11px] text-slate-400">
                  Contextual 3-paragraph executive response synthesized for {lead.contact_name}.
                </p>
              </div>

              <div className="flex items-center gap-2">
                <button
                  onClick={handleCopy}
                  className="flex items-center gap-1.5 px-3 py-1.5 bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-medium rounded-lg border border-slate-700 transition"
                >
                  {isCopied ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
                  <span>{isCopied ? 'Copied!' : 'Copy to Clipboard'}</span>
                </button>

                <button
                  onClick={handleSaveDraft}
                  disabled={isSavingDraft}
                  className="flex items-center gap-1.5 px-3 py-1.5 bg-emerald-600 hover:bg-emerald-500 text-white text-xs font-semibold rounded-lg shadow-sm transition"
                >
                  <Send className="w-3.5 h-3.5" />
                  <span>{isSavingDraft ? 'Saving...' : 'Save Draft'}</span>
                </button>
              </div>
            </div>

            <textarea
              value={draftText}
              onChange={(e) => setDraftText(e.target.value)}
              rows={14}
              className="w-full p-3.5 bg-slate-950 border border-slate-800 rounded-lg text-xs text-slate-200 font-sans leading-relaxed focus:outline-none focus:border-emerald-500/60"
            />
          </div>
        )}

        {/* TAB 3: CRM DISPATCH LOGS */}
        {activeTab === 'crm' && (
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <h4 className="text-xs font-semibold text-white">CRM Synchronization Operations</h4>
              <button
                onClick={handleTriggerResync}
                disabled={isResyncing}
                className="flex items-center gap-1.5 px-3 py-1 bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs rounded border border-slate-700 transition"
              >
                <RotateCw className={`w-3 h-3 ${isResyncing ? 'animate-spin text-emerald-400' : ''}`} />
                <span>Re-sync to CRMs</span>
              </button>
            </div>

            {syncLogs.length === 0 ? (
              <div className="py-8 text-center text-xs text-slate-400">
                No dispatch logs recorded for this lead.
              </div>
            ) : (
              syncLogs.map((log) => (
                <div key={log.id} className="bg-slate-950 border border-slate-800 rounded-lg p-3 space-y-2">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <span className="font-semibold text-xs text-slate-100 uppercase">
                        {log.crm_type}
                      </span>
                      <span className="text-[11px] font-mono text-slate-400">{log.action}</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <span
                        className={`text-[11px] px-2 py-0.5 rounded font-medium ${
                          log.status === 'success'
                            ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20'
                            : 'bg-rose-500/10 text-rose-400 border border-rose-500/20'
                        }`}
                      >
                        {log.status.toUpperCase()} ({log.status_code})
                      </span>
                      <span className="text-[11px] text-slate-400 font-mono">{log.duration_ms}ms</span>
                    </div>
                  </div>

                  {log.response_payload && (
                    <details className="text-[11px]">
                      <summary className="cursor-pointer text-slate-400 hover:text-slate-300">
                        View API Response Payload
                      </summary>
                      <pre className="mt-1.5 p-2 bg-slate-900 text-slate-300 rounded font-mono overflow-x-auto text-[10px]">
                        {log.response_payload}
                      </pre>
                    </details>
                  )}
                </div>
              ))
            )}
          </div>
        )}

        {/* TAB 4: RAW PAYLOAD */}
        {activeTab === 'raw' && (
          <div className="space-y-2">
            <h4 className="text-xs font-semibold text-slate-300 uppercase tracking-wide">
              Raw Inbound Message & Attachments
            </h4>
            <pre className="p-3 bg-slate-950 border border-slate-800 rounded-lg text-slate-300 font-mono text-[11px] overflow-x-auto whitespace-pre-wrap leading-relaxed max-h-[460px]">
              {lead.raw_payload || lead.parsed_body}
            </pre>
          </div>
        )}

        {/* TAB 5: AUDIT TRAIL */}
        {activeTab === 'audit' && (
          <div className="space-y-2">
            <h4 className="text-xs font-semibold text-slate-300 uppercase tracking-wide">
              Pipeline Execution Lifecycle Events
            </h4>
            <div className="space-y-1.5">
              {auditEvents.map((ev) => (
                <div
                  key={ev.id}
                  className="bg-slate-950/70 border border-slate-800/80 rounded-lg p-2.5 flex items-start justify-between text-xs"
                >
                  <div>
                    <div className="flex items-center gap-2">
                      <span className="font-semibold text-emerald-400 font-mono text-[11px]">
                        [{ev.stage.toUpperCase()}]
                      </span>
                      <span className="text-slate-200">{ev.event}</span>
                    </div>
                    {ev.details && <p className="text-[11px] text-slate-400 mt-0.5">{ev.details}</p>}
                  </div>
                  <span className="text-[10px] text-slate-400 font-mono">
                    {new Date(ev.created_at).toLocaleTimeString()}
                  </span>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  );
};
