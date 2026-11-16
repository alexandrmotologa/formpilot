import React, { useState, useEffect } from 'react';
import { Header } from '../components/Header';
import { StatsBar } from '../components/StatsBar';
import { LiveLeadFeed } from '../components/LiveLeadFeed';
import { LeadDetailDrawer } from '../components/LeadDetailDrawer';
import { InboundSimulator } from '../components/InboundSimulator';
import { AccountsCacheModal } from '../components/AccountsCacheModal';
import { SettingsModal } from '../components/SettingsModal';
import { WebhookGatewayModal } from '../components/WebhookGatewayModal';
import { Lead, StatsSummary, HealthInfo, AccountRecord, SyncLog, AuditEvent } from '../types';

export const Dashboard: React.FC = () => {
  const [leads, setLeads] = useState<Lead[]>([]);
  const [selectedLead, setSelectedLead] = useState<Lead | null>(null);
  const [selectedLeadLogs, setSelectedLeadLogs] = useState<SyncLog[]>([]);
  const [selectedLeadAudit, setSelectedLeadAudit] = useState<AuditEvent[]>([]);
  const [stats, setStats] = useState<StatsSummary | null>(null);
  const [health, setHealth] = useState<HealthInfo | null>(null);
  const [accounts, setAccounts] = useState<AccountRecord[]>([]);
  const [sseConnected, setSseConnected] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);

  // Modals
  const [isSimulatorOpen, setIsSimulatorOpen] = useState(false);
  const [isAccountsOpen, setIsAccountsOpen] = useState(false);
  const [isSettingsOpen, setIsSettingsOpen] = useState(false);
  const [isWebhooksOpen, setIsWebhooksOpen] = useState(false);
  const [mobileView, setMobileView] = useState<'feed' | 'detail'>('feed');

  const fetchStats = async () => {
    try {
      const res = await fetch('/api/v1/stats');
      if (res.ok) {
        const data = await res.json();
        setStats(data);
      }
    } catch {
      // offline
    }
  };

  const fetchHealth = async () => {
    try {
      const res = await fetch('/api/v1/health');
      if (res.ok) {
        const data = await res.json();
        setHealth(data);
      }
    } catch {
      // offline
    }
  };

  const fetchAccounts = async () => {
    try {
      const res = await fetch('/api/v1/accounts');
      if (res.ok) {
        const data = await res.json();
        setAccounts(data);
      }
    } catch {
      // offline
    }
  };

  const fetchLeads = async () => {
    try {
      const res = await fetch('/api/v1/leads');
      if (res.ok) {
        const data = await res.json();
        setLeads(data || []);
        if (data && data.length > 0 && !selectedLead) {
          handleSelectLead(data[0]);
        }
      }
    } catch {
      // offline
    } finally {
      setIsLoading(false);
    }
  };

  const handleSelectLead = async (lead: Lead) => {
    setSelectedLead(lead);
    setMobileView('detail');
    try {
      const res = await fetch(`/api/v1/leads/${lead.id}`);
      if (res.ok) {
        const data = await res.json();
        setSelectedLeadLogs(data.sync_logs || []);
        setSelectedLeadAudit(data.audit_events || []);
      }
    } catch {
      setSelectedLeadLogs([]);
      setSelectedLeadAudit([]);
    }
  };

  const handleRefresh = async () => {
    setIsRefreshing(true);
    await Promise.all([fetchLeads(), fetchStats(), fetchHealth(), fetchAccounts()]);
    setIsRefreshing(false);
  };

  // Initial load
  useEffect(() => {
    handleRefresh();
  }, []);

  // SSE setup
  useEffect(() => {
    const eventSource = new EventSource('/api/v1/events');

    eventSource.onopen = () => {
      setSseConnected(true);
    };

    eventSource.onerror = () => {
      setSseConnected(false);
    };

    eventSource.addEventListener('lead.processed', (e: MessageEvent) => {
      try {
        const newLead: Lead = JSON.parse(e.data);
        setLeads((prev) => {
          const index = prev.findIndex((l) => l.id === newLead.id);
          if (index >= 0) {
            const updated = [...prev];
            updated[index] = newLead;
            return updated;
          }
          return [newLead, ...prev];
        });

        // Auto select if inspecting or first
        setSelectedLead((current) => {
          if (!current || current.id === newLead.id) {
            handleSelectLead(newLead);
            return newLead;
          }
          return current;
        });

        fetchStats();
        fetchAccounts();
      } catch {
        // ignore
      }
    });

    eventSource.addEventListener('lead.draft_updated', (e: MessageEvent) => {
      try {
        const updatedLead: Lead = JSON.parse(e.data);
        setLeads((prev) =>
          prev.map((l) => (l.id === updatedLead.id ? updatedLead : l))
        );
        setSelectedLead((current) =>
          current && current.id === updatedLead.id ? updatedLead : current
        );
      } catch {
        // ignore
      }
    });

    eventSource.addEventListener('stats.updated', () => {
      fetchLeads();
      fetchStats();
    });

    return () => {
      eventSource.close();
    };
  }, []);

  const handleUpdateDraft = async (leadId: string, newDraft: string) => {
    const res = await fetch(`/api/v1/leads/${leadId}/draft`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ draft: newDraft }),
    });
    if (res.ok) {
      const updated = await res.json();
      setSelectedLead(updated);
      setLeads((prev) => prev.map((l) => (l.id === updated.id ? updated : l)));
    }
  };

  const handleRegenerateDraft = async (leadId: string, tone: string) => {
    const res = await fetch(`/api/v1/leads/${leadId}/regenerate-draft`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ tone }),
    });
    if (res.ok) {
      const updated = await res.json();
      setSelectedLead(updated);
      setLeads((prev) => prev.map((l) => (l.id === updated.id ? updated : l)));
    }
  };

  const handleSendEmail = async (leadId: string) => {
    const res = await fetch(`/api/v1/leads/${leadId}/send-email`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({}),
    });
    if (res.ok) {
      const leadRes = await fetch(`/api/v1/leads/${leadId}`);
      if (leadRes.ok) {
        const data = await leadRes.json();
        handleSelectLead(data.lead);
      }
      fetchStats();
    }
  };

  const handleOverrideDedup = async (leadId: string, action: string) => {
    const res = await fetch(`/api/v1/leads/${leadId}/override-dedup`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ action }),
    });
    if (res.ok) {
      const updated = await res.json();
      handleSelectLead(updated);
      fetchStats();
    }
  };

  const handleResync = async (leadId: string) => {
    const res = await fetch(`/api/v1/leads/${leadId}/sync`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
    });
    if (res.ok) {
      const updated = await res.json();
      handleSelectLead(updated);
      fetchStats();
    }
  };

  const handleExport = (format: 'csv' | 'json') => {
    window.open(`/api/v1/leads/export?format=${format}`, '_blank');
  };

  const handleSimulate = async (preset: string): Promise<Lead> => {
    const res = await fetch('/api/v1/simulate', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ preset }),
    });
    const data = await res.json();
    handleSelectLead(data);
    return data;
  };

  const handleCustomSubmit = async (sender: string, subject: string, body: string): Promise<Lead> => {
    const res = await fetch('/api/v1/webhooks/generic', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ sender, subject, body, source: 'cockpit_simulator' }),
    });
    const result = await res.json();
    const leadRes = await fetch(`/api/v1/leads/${result.lead_id}`);
    const fullLeadData = await leadRes.json();
    handleSelectLead(fullLeadData.lead);
    return fullLeadData.lead;
  };

  return (
    <div className="flex flex-col h-screen bg-slate-950 text-slate-100 overflow-hidden font-sans">
      <Header
        health={health}
        sseConnected={sseConnected}
        onOpenSimulator={() => setIsSimulatorOpen(true)}
        onOpenAccounts={() => setIsAccountsOpen(true)}
        onOpenSettings={() => setIsSettingsOpen(true)}
        onOpenWebhooks={() => setIsWebhooksOpen(true)}
        onExport={handleExport}
        onRefresh={handleRefresh}
        isRefreshing={isRefreshing}
      />

      <main className="flex-1 overflow-hidden p-3 sm:p-5 lg:p-6 flex flex-col max-w-7xl mx-auto w-full">
        <StatsBar stats={stats} />

        {/* Mobile View Switcher (visible only on < lg screens) */}
        <div className="flex lg:hidden items-center justify-between mb-3 bg-slate-900/90 border border-slate-800 rounded-lg p-1 shrink-0">
          <button
            onClick={() => setMobileView('feed')}
            className={`flex-1 py-1.5 text-xs font-semibold rounded-md transition ${
              mobileView === 'feed'
                ? 'bg-slate-800 text-white shadow'
                : 'text-slate-400 hover:text-slate-200'
            }`}
          >
            Inbound Feed ({leads.length})
          </button>
          <button
            onClick={() => setMobileView('detail')}
            disabled={!selectedLead}
            className={`flex-1 py-1.5 text-xs font-semibold rounded-md transition ${
              mobileView === 'detail'
                ? 'bg-slate-800 text-white shadow'
                : 'text-slate-400 hover:text-slate-200 disabled:opacity-40 disabled:cursor-not-allowed'
            }`}
          >
            {selectedLead ? selectedLead.company_name || 'Lead Details' : 'Lead Details'}
          </button>
        </div>

        {/* Two-column responsive Cockpit workspace */}
        <div className="flex-1 grid grid-cols-1 lg:grid-cols-12 gap-5 min-h-0">
          {/* Left: Real-time Inbound Feed (5 cols) */}
          <div className={`lg:col-span-5 h-full overflow-hidden ${mobileView === 'feed' ? 'block' : 'hidden lg:block'}`}>
            <LiveLeadFeed
              leads={leads}
              selectedLeadId={selectedLead?.id || null}
              onSelectLead={handleSelectLead}
              isLoading={isLoading}
            />
          </div>

          {/* Right: Lead Detail Drawer & Outreach Studio (7 cols) */}
          <div className={`lg:col-span-7 h-full overflow-hidden ${mobileView === 'detail' ? 'block' : 'hidden lg:block'}`}>
            <LeadDetailDrawer
              lead={selectedLead}
              syncLogs={selectedLeadLogs}
              auditEvents={selectedLeadAudit}
              onClose={() => {
                setSelectedLead(null);
                setMobileView('feed');
              }}
              onUpdateDraft={handleUpdateDraft}
              onResync={handleResync}
              onRegenerateDraft={handleRegenerateDraft}
              onSendEmail={handleSendEmail}
              onOverrideDedup={handleOverrideDedup}
            />
          </div>
        </div>
      </main>

      <InboundSimulator
        isOpen={isSimulatorOpen}
        onClose={() => setIsSimulatorOpen(false)}
        onSimulate={handleSimulate}
        onCustomSubmit={handleCustomSubmit}
      />

      <AccountsCacheModal
        isOpen={isAccountsOpen}
        onClose={() => setIsAccountsOpen(false)}
        accounts={accounts}
      />

      <SettingsModal
        isOpen={isSettingsOpen}
        onClose={() => setIsSettingsOpen(false)}
        onRefreshStats={fetchStats}
      />

      <WebhookGatewayModal
        isOpen={isWebhooksOpen}
        onClose={() => setIsWebhooksOpen(false)}
      />
    </div>
  );
};
