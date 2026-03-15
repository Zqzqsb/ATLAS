#!/usr/bin/env python3
"""
Generate ADDON tables for tpch_enterprise (appended to existing 38 core tables).
Only outputs USE + CREATE TABLE + INSERT for NEW tables. No DROP, no re-create.

Usage:
    python3 scripts/gen_tpch_addon.py > /tmp/tpch_addon.sql
    docker exec -i lucid-mariadb mariadb -uroot -p... < /tmp/tpch_addon.sql
"""
import random, re
random.seed(42)

# ──── Column generator ────────────────────────────────────────
def gen_val(col_type, col_name, row_idx):
    ct, cn = col_type.upper(), col_name.lower()
    if "AUTO_INCREMENT" in ct:
        return None
    len_m = re.search(r'\((\d+)\)', ct)
    max_len = int(len_m.group(1)) if len_m else 50

    if "TINYINT" in ct:
        return str(random.choice([0,1]))
    if "INT" in ct:
        if "year" in cn: return str(2024 + row_idx % 3)
        if "count" in cn or "qty" in cn: return str(random.randint(1,200))
        return str(random.randint(1,100))
    if "DECIMAL" in ct:
        if "pct" in cn or "rate" in cn: return f"{random.uniform(1,80):.2f}"
        return f"{random.uniform(10,9999):.2f}"
    if "DATETIME" in ct or "TIMESTAMP" in ct:
        d = random.randint(0,364)
        h = random.randint(0,23)
        return f"'2025-{1+d//30:02d}-{1+d%28:02d} {h:02d}:00:00'"
    if "DATE" in ct:
        d = random.randint(0,364)
        return f"'2025-{1+d//30:02d}-{1+d%28:02d}'"
    if "TEXT" in ct:
        return f"'Sample data row {row_idx+1}'"
    # VARCHAR / CHAR
    if "status" in cn: v = random.choice(["active","pending","done"])
    elif "email" in cn: v = f"u{row_idx+1}@ex.com"
    elif "name" in cn or "title" in cn:
        w = ["Alpha","Beta","Gamma","Delta","Eps"][row_idx%5]
        v = f"{w}-{row_idx+1}"
    elif "type" in cn or "category" in cn: v = f"type_{chr(65+row_idx%5)}"
    elif "code" in cn: v = f"{cn[:2].upper()}{1000+row_idx}"
    elif "priority" in cn or "severity" in cn: v = random.choice(["low","normal","high"])
    elif "method" in cn or "channel" in cn: v = random.choice(["online","phone","email"])
    elif "currency" in cn: v = random.choice(["USD","EUR","GBP"])
    elif "country" in cn: v = random.choice(["USA","UK","DE"])
    elif "level" in cn or "grade" in cn: v = f"L{row_idx+1}"
    elif "phone" in cn: v = f"+1-555-{random.randint(1000,9999)}"
    else: v = f"{cn[:6]}_{row_idx+1}"
    return f"'{v[:max_len]}'"

def make_cols(name, fk_ref=None):
    """Auto-generate columns for a table based on name patterns.
    fk_ref: if set, (col_name, ref_table, ref_col) to add a FK column."""
    cols = [("id","INT PRIMARY KEY AUTO_INCREMENT","ID")]
    cols.append(("name","VARCHAR(120)","Name"))
    # Add FK column if this table references a hub table
    if fk_ref:
        fk_col, ref_tbl, ref_col = fk_ref
        cols.append((fk_col, f"INT COMMENT 'FK to {ref_tbl}'", f"Ref {ref_tbl}"))
    n = name.lower()
    if any(k in n for k in ["plan","schedule","forecast","budget","target"]):
        cols += [("fiscal_year","INT","Year"),("planned_value","DECIMAL(14,2)","Planned"),("actual_value","DECIMAL(14,2) DEFAULT 0","Actual")]
    elif any(k in n for k in ["record","log","entry","event","transaction"]):
        cols += [("event_date","DATETIME DEFAULT CURRENT_TIMESTAMP","Date"),("description","TEXT","Description"),("created_by","VARCHAR(80)","Creator")]
    elif any(k in n for k in ["metric","score","kpi","measurement"]):
        cols += [("metric_date","DATE","Date"),("value","DECIMAL(14,4)","Value"),("target","DECIMAL(14,4)","Target")]
    else:
        cols += [("description","VARCHAR(300)","Description"),("category","VARCHAR(50)","Category"),("created_at","DATETIME DEFAULT CURRENT_TIMESTAMP","Created")]
    cols.append(("status","VARCHAR(20) DEFAULT 'active'","Status"))
    return cols

def emit_table(prefix, suffix, comment, cols, rows=3, fk_constraint=None):
    full = f"{prefix}_{suffix}"
    # CREATE TABLE
    cdefs = []
    for cn, ct, cc in cols:
        cdefs.append(f"    {cn:28s} {ct} COMMENT '{cc}'")
    if fk_constraint:
        fk_col, ref_tbl, ref_col = fk_constraint
        cdefs.append(f"    FOREIGN KEY ({fk_col}) REFERENCES {ref_tbl}({ref_col})")
    print(f"CREATE TABLE IF NOT EXISTS {full} (\n" + ",\n".join(cdefs) + f"\n) COMMENT='{comment}';")
    # INSERT
    insertable = [(cn,ct) for cn,ct,_ in cols if "AUTO_INCREMENT" not in ct.upper()]
    if not insertable:
        print()
        return
    cnames = [cn for cn,_ in insertable]
    vals = []
    for i in range(rows):
        rv = []
        for cn,ct in insertable:
            v = gen_val(ct, cn, i)
            if v is not None: rv.append(v)
        vals.append(f"({', '.join(rv)})")
    print(f"INSERT IGNORE INTO {full} ({', '.join(cnames)}) VALUES\n" + ",\n".join(vals) + ";\n")


# ──── Domain definitions ──────────────────────────────────────
DOMAINS = {
    "hr": ["departments","employees","positions","recruitment_requisitions","candidates","interviews",
           "offer_letters","training_courses","training_enrollments","performance_goals","benefits_plans",
           "benefits_enrollments","leave_requests","leave_balances","timesheets","expense_reports",
           "expense_items","employee_documents","org_chart_history","salary_bands","compensation_history",
           "skills","employee_skills","employee_certifications","disciplinary_actions","employee_surveys",
           "succession_plans","onboarding_tasks","exit_interviews","payroll_runs","payroll_details",
           "workforce_forecast","emergency_contacts","grievances","workplace_incidents"],
    "fin": ["chart_of_accounts","journal_entries","journal_lines","fiscal_periods","vendor_master",
            "purchase_orders","po_line_items","ap_invoices","ap_payments","ar_invoices","ar_receipts",
            "bank_accounts","bank_transactions","fixed_assets","depreciation_schedules","budgets",
            "cost_centers","exchange_rates","tax_jurisdictions","intercompany_txns","credit_memos",
            "financial_reports","audit_findings","cash_flow_forecasts","revenue_recognition"],
    "scm": ["procurement_requests","supplier_evaluations","goods_receipts","returns_to_vendor",
            "shipping_carriers","shipment_tracking","demand_forecasts","safety_stock_levels",
            "customs_declarations","inbound_deliveries","outbound_shipments","supplier_contracts",
            "material_requirements","freight_invoices"],
    "crm": ["customer_segments","contact_interactions","support_tickets","ticket_responses",
            "loyalty_programs","loyalty_transactions","feedback_forms","nps_scores",
            "customer_addresses","referral_programs"],
    "mfg": ["bill_of_materials","work_orders","work_centers","production_schedules",
            "quality_inspections","scrap_records","equipment_master","maintenance_orders",
            "tooling_inventory","production_kpis"],
    "sales": ["price_lists","price_list_items","sales_quotations","quote_line_items",
              "sales_orders","sales_order_items","sales_territories","sales_commissions",
              "sales_returns","sales_targets"],
    "proj": ["tasks","milestones","resource_allocations","risk_register","project_budgets",
             "status_reports","project_documents","stakeholders","change_requests","project_phases",
             "dependencies","issue_log","lessons_learned","meeting_minutes","deliverables",
             "work_packages","gantt_schedules","project_portfolios","program_master","project_templates",
             "approval_workflows","time_entries","resource_skills","project_kpis","sprint_backlogs"],
    "qa": ["inspection_plans","test_cases","test_runs","defect_reports","corrective_actions",
           "preventive_actions","audit_schedules","audit_findings","nonconformance_reports",
           "control_plans","measurement_systems","calibration_records","spc_data","process_capability",
           "reliability_tests","environmental_tests","supplier_quality","incoming_inspections",
           "final_inspections","customer_complaints","root_cause_analysis","capa_tracking",
           "quality_costs","quality_objectives","document_controls"],
    "it": ["it_assets","software_licenses","hardware_inventory","help_desk_tickets","ticket_comments",
           "change_requests","change_approvals","sla_definitions","sla_metrics","network_devices",
           "server_inventory","backup_schedules","backup_logs","security_incidents","vulnerability_scans",
           "patch_management","access_requests","it_projects","service_catalog","capacity_metrics",
           "configuration_items","release_management","incident_escalations","knowledge_articles","it_budgets"],
    "legal": ["contracts","contract_amendments","legal_cases","case_documents","intellectual_property",
              "trademark_registrations","patent_filings","regulatory_requirements","compliance_checklist",
              "compliance_audits","legal_holds","litigation_matters","settlement_records","legal_fees",
              "attorney_assignments","nda_register","corporate_filings","policy_documents",
              "regulatory_fines","legal_calendar"],
    "mktg": ["marketing_campaigns","campaign_budgets","campaign_performance","email_campaigns",
             "email_templates","social_media_posts","social_media_metrics","content_calendar",
             "content_assets","seo_keywords","seo_rankings","ppc_campaigns","ppc_ad_groups",
             "landing_pages","conversion_tracking","webinar_events","webinar_registrations",
             "trade_shows","brand_guidelines","market_research","competitor_analysis",
             "press_releases","influencer_partnerships","media_buys","affiliate_programs"],
    "wms": ["warehouse_zones","storage_bins","bin_assignments","pick_lists","pick_tasks",
            "pack_stations","packing_records","receiving_docks","putaway_tasks","cycle_count_plans",
            "cycle_count_results","wave_planning","replenishment_tasks","cross_dock_orders",
            "kitting_orders","kit_components","returns_processing","quarantine_areas",
            "temperature_logs","warehouse_kpis","labor_tracking","forklift_assignments",
            "yard_management","dock_schedules","barcode_labels"],
    "fleet": ["vehicles","vehicle_maintenance","drivers","driver_licenses","routes","route_stops",
              "fuel_records","fuel_cards","trip_logs","vehicle_inspections","accident_reports",
              "insurance_policies","vehicle_assignments","toll_records","parking_permits",
              "gps_tracking","vehicle_leases","tire_records","emissions_tests","fleet_costs"],
    "esg": ["carbon_emissions","emission_sources","energy_consumption","energy_sources","water_usage",
            "waste_generation","waste_disposal","recycling_records","sustainability_goals",
            "sustainability_metrics","social_impact_projects","community_investments","diversity_metrics",
            "safety_metrics","supply_chain_ethics","environmental_audits","carbon_offsets",
            "renewable_energy_certs","esg_reports","stakeholder_engagement"],
    "rnd": ["research_projects","experiments","experiment_results","prototypes","prototype_tests",
            "patent_applications","patent_citations","publications","research_grants","grant_milestones",
            "lab_equipment","lab_bookings","reagent_inventory","clinical_trials","trial_participants",
            "innovation_ideas","idea_evaluations","technology_roadmaps","research_partnerships",
            "technical_reviews"],
    "bi": ["kpi_definitions","kpi_measurements","dashboard_definitions","dashboard_widgets",
           "data_sources","etl_jobs","etl_job_runs","data_quality_rules","data_quality_scores",
           "report_definitions","report_schedules","report_distributions","data_catalogs",
           "data_lineage","metric_alerts","alert_notifications","dimension_tables","fact_tables",
           "cube_definitions","data_governance_policies"],
    "acct": ["tax_filings","tax_schedules","withholding_records","expense_categories","expense_policies",
             "travel_requests","travel_itineraries","mileage_claims","petty_cash_funds",
             "petty_cash_transactions","asset_disposals","lease_agreements","lease_payments",
             "loan_records","loan_payments","investment_portfolios","investment_transactions",
             "dividend_records","insurance_premiums","insurance_claims"],
    "proc": ["rfq_documents","rfq_responses","bid_evaluations","framework_agreements","blanket_orders",
             "blanket_releases","catalog_items","catalog_prices","vendor_onboarding","vendor_documents",
             "vendor_contacts","vendor_performance_kpis","sourcing_events","sourcing_awards",
             "contract_milestones","contract_deliverables","spend_analysis","commodity_codes",
             "approved_vendor_list","purchase_req_lines"],
    "cust": ["customer_tiers","tier_benefits","customer_preferences","comm_preferences",
             "subscription_plans","subscription_billing","health_scores","churn_predictions",
             "win_back_campaigns","customer_journeys","touchpoint_analysis","customer_360_views",
             "account_teams","account_plans","strategic_accounts","customer_portals",
             "portal_activity_logs","self_service_tickets","knowledge_base_articles","faq_categories"],
    "plnt": ["plant_master","plant_areas","production_lines","line_stations","station_assignments",
             "shift_schedules","shift_assignments","downtime_events","downtime_reasons","oee_metrics",
             "takt_time_records","andon_alerts","kanban_cards","kanban_boards","visual_mgmt_boards",
             "gemba_walk_records","kaizen_suggestions","suggestion_reviews","ci_projects","ci_results"],
    "ops": ["batch_jobs","batch_steps","mq_topics","mq_subscriptions","api_endpoints","api_usage_logs",
            "webhook_configs","webhook_deliveries","doc_templates","doc_versions","workflow_defs",
            "workflow_instances","workflow_steps","notif_templates","notif_channels","scheduled_tasks",
            "task_exec_logs","feature_flags","ab_experiments","ab_results","geo_regions","geo_cities",
            "timezone_defs","currency_master","language_packs","translation_keys","user_preferences",
            "system_params","health_checks","perf_baselines","capacity_plans","dr_plans",
            "backup_policies","data_class_rules","key_rotations","cert_mgmt","dns_records",
            "lb_configs","container_deploys","mesh_policies","rate_limits","oauth_clients",
            "oauth_tokens","saml_providers","mfa_configs","login_audit","permission_sets",
            "role_assignments","resource_quotas","cost_tags","tag_defs","tag_assignments",
            "event_subs","event_handlers","cron_schedules","cron_history","file_buckets",
            "file_metadata","img_proc_jobs","pdf_gen_jobs","email_logs","sms_logs",
            "push_notif_logs","in_app_msgs","activity_feed","feed_comments",
            "global_settings","tenant_configs","feature_entitlements","usage_metering"],
}

DOMAIN_LABELS = {
    "hr":"HR","fin":"Finance","scm":"Supply Chain","crm":"CRM","mfg":"Manufacturing",
    "sales":"Sales","proj":"Project Mgmt","qa":"Quality","it":"IT","legal":"Legal",
    "mktg":"Marketing","wms":"Warehouse","fleet":"Fleet","esg":"ESG","rnd":"R&D",
    "bi":"BI","acct":"Accounting","proc":"Procurement","cust":"Customer Success",
    "plnt":"Plant Ops","ops":"Operations",
}

def main():
    print("USE tpch_enterprise;\n")
    total = 0
    for prefix, tables in DOMAINS.items():
        label = DOMAIN_LABELS.get(prefix, prefix)
        print(f"-- === {label} ({len(tables)} tables) ===\n")
        # First table in each domain is the "hub" — no FK
        hub_table = f"{prefix}_{tables[0]}"
        for idx, t in enumerate(tables):
            if idx == 0:
                # Hub table: no FK reference
                cols = make_cols(t)
                comment = f"{label} — {t.replace('_',' ').title()}"
                emit_table(prefix, t, comment, cols, rows=random.randint(2,4))
            else:
                # Spoke tables: FK to hub table's id column
                fk_col = f"{tables[0]}_id"
                fk_ref = (fk_col, hub_table, "id")
                cols = make_cols(t, fk_ref=fk_ref)
                comment = f"{label} — {t.replace('_',' ').title()}"
                emit_table(prefix, t, comment, cols, rows=random.randint(2,4),
                           fk_constraint=(fk_col, hub_table, "id"))
            total += 1
    import sys
    print(f"-- Addon tables generated: {total}", file=sys.stderr)

if __name__ == "__main__":
    main()
