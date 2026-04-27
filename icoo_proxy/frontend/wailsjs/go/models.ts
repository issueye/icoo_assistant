export namespace api {
	
	export class EndpointView {
	    id: string;
	    path: string;
	    protocol: string;
	    description: string;
	    enabled: boolean;
	    built_in: boolean;
	    updated_at: string;
	    created_at: string;
	
	    static createFrom(source: any = {}) {
	        return new EndpointView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.path = source["path"];
	        this.protocol = source["protocol"];
	        this.description = source["description"];
	        this.enabled = source["enabled"];
	        this.built_in = source["built_in"];
	        this.updated_at = source["updated_at"];
	        this.created_at = source["created_at"];
	    }
	}
	export class RequestView {
	    request_id: string;
	    downstream: string;
	    upstream: string;
	    model: string;
	    status_code: number;
	    duration_ms: number;
	    error?: string;
	    created_at: string;
	
	    static createFrom(source: any = {}) {
	        return new RequestView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.request_id = source["request_id"];
	        this.downstream = source["downstream"];
	        this.upstream = source["upstream"];
	        this.model = source["model"];
	        this.status_code = source["status_code"];
	        this.duration_ms = source["duration_ms"];
	        this.error = source["error"];
	        this.created_at = source["created_at"];
	    }
	}
	export class RoutePolicyView {
	    id: string;
	    downstream_protocol: string;
	    supplier_id: string;
	    supplier_name: string;
	    upstream_protocol: string;
	    enabled: boolean;
	    updated_at: string;
	    created_at: string;
	
	    static createFrom(source: any = {}) {
	        return new RoutePolicyView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.downstream_protocol = source["downstream_protocol"];
	        this.supplier_id = source["supplier_id"];
	        this.supplier_name = source["supplier_name"];
	        this.upstream_protocol = source["upstream_protocol"];
	        this.enabled = source["enabled"];
	        this.updated_at = source["updated_at"];
	        this.created_at = source["created_at"];
	    }
	}
	export class RouteView {
	    name: string;
	    upstream: string;
	    model: string;
	
	    static createFrom(source: any = {}) {
	        return new RouteView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.upstream = source["upstream"];
	        this.model = source["model"];
	    }
	}
	export class UpstreamView {
	    protocol: string;
	    base_url?: string;
	    configured: boolean;
	
	    static createFrom(source: any = {}) {
	        return new UpstreamView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.protocol = source["protocol"];
	        this.base_url = source["base_url"];
	        this.configured = source["configured"];
	    }
	}
	export class State {
	    service: string;
	    version: string;
	    running: boolean;
	    listen_addr?: string;
	    proxy_url?: string;
	    last_error?: string;
	    auth_required: boolean;
	    auth_key_count: number;
	    allow_unauthenticated_local: boolean;
	    supported_paths: string[];
	    defaults: RouteView[];
	    aliases: RouteView[];
	    upstreams: UpstreamView[];
	    endpoints: EndpointView[];
	    route_policies: RoutePolicyView[];
	    recent_requests: RequestView[];
	    notes: string[];
	    checks: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new State(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.service = source["service"];
	        this.version = source["version"];
	        this.running = source["running"];
	        this.listen_addr = source["listen_addr"];
	        this.proxy_url = source["proxy_url"];
	        this.last_error = source["last_error"];
	        this.auth_required = source["auth_required"];
	        this.auth_key_count = source["auth_key_count"];
	        this.allow_unauthenticated_local = source["allow_unauthenticated_local"];
	        this.supported_paths = source["supported_paths"];
	        this.defaults = this.convertValues(source["defaults"], RouteView);
	        this.aliases = this.convertValues(source["aliases"], RouteView);
	        this.upstreams = this.convertValues(source["upstreams"], UpstreamView);
	        this.endpoints = this.convertValues(source["endpoints"], EndpointView);
	        this.route_policies = this.convertValues(source["route_policies"], RoutePolicyView);
	        this.recent_requests = this.convertValues(source["recent_requests"], RequestView);
	        this.notes = source["notes"];
	        this.checks = source["checks"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace authkey {
	
	export class Record {
	    id: string;
	    name: string;
	    secret_masked: string;
	    enabled: boolean;
	    description: string;
	    updated_at: string;
	    created_at: string;
	
	    static createFrom(source: any = {}) {
	        return new Record(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.secret_masked = source["secret_masked"];
	        this.enabled = source["enabled"];
	        this.description = source["description"];
	        this.updated_at = source["updated_at"];
	        this.created_at = source["created_at"];
	    }
	}
	export class UpsertInput {
	    id: string;
	    name: string;
	    secret: string;
	    enabled: boolean;
	    description: string;
	
	    static createFrom(source: any = {}) {
	        return new UpsertInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.secret = source["secret"];
	        this.enabled = source["enabled"];
	        this.description = source["description"];
	    }
	}

}

export namespace endpoint {
	
	export class Record {
	    id: string;
	    path: string;
	    protocol: string;
	    description: string;
	    enabled: boolean;
	    built_in: boolean;
	    updated_at: string;
	    created_at: string;
	
	    static createFrom(source: any = {}) {
	        return new Record(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.path = source["path"];
	        this.protocol = source["protocol"];
	        this.description = source["description"];
	        this.enabled = source["enabled"];
	        this.built_in = source["built_in"];
	        this.updated_at = source["updated_at"];
	        this.created_at = source["created_at"];
	    }
	}
	export class UpsertInput {
	    id: string;
	    path: string;
	    protocol: string;
	    description: string;
	    enabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new UpsertInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.path = source["path"];
	        this.protocol = source["protocol"];
	        this.description = source["description"];
	        this.enabled = source["enabled"];
	    }
	}

}

export namespace modelalias {
	
	export class Record {
	    id: string;
	    name: string;
	    supplier_id: string;
	    supplier_name: string;
	    upstream_protocol: string;
	    model: string;
	    enabled: boolean;
	    updated_at: string;
	    created_at: string;
	
	    static createFrom(source: any = {}) {
	        return new Record(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.supplier_id = source["supplier_id"];
	        this.supplier_name = source["supplier_name"];
	        this.upstream_protocol = source["upstream_protocol"];
	        this.model = source["model"];
	        this.enabled = source["enabled"];
	        this.updated_at = source["updated_at"];
	        this.created_at = source["created_at"];
	    }
	}
	export class UpsertInput {
	    id: string;
	    name: string;
	    supplier_id: string;
	    model: string;
	    enabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new UpsertInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.supplier_id = source["supplier_id"];
	        this.model = source["model"];
	        this.enabled = source["enabled"];
	    }
	}

}

export namespace projectsettings {
	
	export class Values {
	    proxy_host: string;
	    proxy_port: number;
	    proxy_read_timeout_seconds: number;
	    proxy_write_timeout_seconds: number;
	    proxy_shutdown_timeout_seconds: number;
	    proxy_chain_log_path: string;
	    proxy_chain_log_bodies: boolean;
	    proxy_chain_log_max_body_bytes: number;
	
	    static createFrom(source: any = {}) {
	        return new Values(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.proxy_host = source["proxy_host"];
	        this.proxy_port = source["proxy_port"];
	        this.proxy_read_timeout_seconds = source["proxy_read_timeout_seconds"];
	        this.proxy_write_timeout_seconds = source["proxy_write_timeout_seconds"];
	        this.proxy_shutdown_timeout_seconds = source["proxy_shutdown_timeout_seconds"];
	        this.proxy_chain_log_path = source["proxy_chain_log_path"];
	        this.proxy_chain_log_bodies = source["proxy_chain_log_bodies"];
	        this.proxy_chain_log_max_body_bytes = source["proxy_chain_log_max_body_bytes"];
	    }
	}

}

export namespace routepolicy {
	
	export class Record {
	    id: string;
	    downstream_protocol: string;
	    supplier_id: string;
	    supplier_name: string;
	    upstream_protocol: string;
	    enabled: boolean;
	    updated_at: string;
	    created_at: string;
	
	    static createFrom(source: any = {}) {
	        return new Record(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.downstream_protocol = source["downstream_protocol"];
	        this.supplier_id = source["supplier_id"];
	        this.supplier_name = source["supplier_name"];
	        this.upstream_protocol = source["upstream_protocol"];
	        this.enabled = source["enabled"];
	        this.updated_at = source["updated_at"];
	        this.created_at = source["created_at"];
	    }
	}
	export class UpsertInput {
	    id: string;
	    downstream_protocol: string;
	    supplier_id: string;
	    enabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new UpsertInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.downstream_protocol = source["downstream_protocol"];
	        this.supplier_id = source["supplier_id"];
	        this.enabled = source["enabled"];
	    }
	}

}

export namespace supplier {
	
	export class HealthRecord {
	    supplier_id: string;
	    status: string;
	    message: string;
	    checked_at: string;
	    status_code: number;
	    duration_ms: number;
	    reachable: boolean;
	    protocol: string;
	    base_url: string;
	    supplier_name: string;
	
	    static createFrom(source: any = {}) {
	        return new HealthRecord(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.supplier_id = source["supplier_id"];
	        this.status = source["status"];
	        this.message = source["message"];
	        this.checked_at = source["checked_at"];
	        this.status_code = source["status_code"];
	        this.duration_ms = source["duration_ms"];
	        this.reachable = source["reachable"];
	        this.protocol = source["protocol"];
	        this.base_url = source["base_url"];
	        this.supplier_name = source["supplier_name"];
	    }
	}
	export class Record {
	    id: string;
	    name: string;
	    protocol: string;
	    base_url: string;
	    api_key_masked: string;
	    only_stream: boolean;
	    user_agent: string;
	    enabled: boolean;
	    description: string;
	    models: string[];
	    default_model: string;
	    updated_at: string;
	    created_at: string;
	
	    static createFrom(source: any = {}) {
	        return new Record(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.protocol = source["protocol"];
	        this.base_url = source["base_url"];
	        this.api_key_masked = source["api_key_masked"];
	        this.only_stream = source["only_stream"];
	        this.user_agent = source["user_agent"];
	        this.enabled = source["enabled"];
	        this.description = source["description"];
	        this.models = source["models"];
	        this.default_model = source["default_model"];
	        this.updated_at = source["updated_at"];
	        this.created_at = source["created_at"];
	    }
	}
	export class UpsertInput {
	    id: string;
	    name: string;
	    protocol: string;
	    base_url: string;
	    api_key: string;
	    only_stream: boolean;
	    user_agent: string;
	    enabled: boolean;
	    description: string;
	    models: string;
	    default_model: string;
	
	    static createFrom(source: any = {}) {
	        return new UpsertInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.protocol = source["protocol"];
	        this.base_url = source["base_url"];
	        this.api_key = source["api_key"];
	        this.only_stream = source["only_stream"];
	        this.user_agent = source["user_agent"];
	        this.enabled = source["enabled"];
	        this.description = source["description"];
	        this.models = source["models"];
	        this.default_model = source["default_model"];
	    }
	}

}

export namespace uiprefs {
	
	export class Preferences {
	    theme: string;
	    buttonSize: string;
	
	    static createFrom(source: any = {}) {
	        return new Preferences(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.theme = source["theme"];
	        this.buttonSize = source["buttonSize"];
	    }
	}

}

