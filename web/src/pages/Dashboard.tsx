import React, { useState, useEffect, useCallback } from 'react';
import { Header } from '../components/Header';
import { StatsBar } from '../components/StatsBar';
import { LiveLeadFeed } from '../components/LiveLeadFeed';
import { LeadDetailDrawer } from '../components/LeadDetailDrawer';
import { InboundSimulator } from '../components/InboundSimulator';
import { AccountsCacheModal } from '../components/AccountsCacheModal';
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
  const [isSimulatorOpen, setIsSimulatorOpen] = useState(false);
  const [isAccountsOpen, setIsAccountsOpen] = useState(false);

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

        // Auto select if first or currently inspecting
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
        // ignore parse error
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
    // fetch full lead
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
        onRefresh={handleRefresh}
        isRefreshing={isRefreshing}
      />

      <main className="flex-1 overflow-hidden p-6 flex flex-col max-w-7xl mx-auto w-full">
        <StatsBar stats={stats} />

        {/* Two-column responsive Cockpit workspace */}
        <div className="flex-1 grid grid-cols-1 lg:grid-cols-12 gap-5 min-h-0">
          {/* Left: Real-time Inbound Feed (5 cols) */}
          <div className="lg:col-span-5 h-full overflow-hidden">
            <LiveLeadFeed
              leads={leads}
              selectedLeadId={selectedLead?.id || null}
              onSelectLead={handleSelectLead}
              isLoading={isLoading}
            />
          </div>

          {/* Right: Lead Detail Drawer & Outreach Studio (7 cols) */}
          <div className="lg:col-span-7 h-full overflow-hidden">
            <LeadDetailDrawer
              lead={selectedLead}
              syncLogs={selectedLeadLogs}
              auditEvents={selectedLeadAudit}
              onClose={() => setSelectedLead(null)}
              onUpdateDraft={handleUpdateDraft}
              onResync={handleResync}
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
    </div>
  );
};
