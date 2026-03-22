export interface PublicStatus {
  version: string;
  start_time: number;
  email_verification: boolean;
  github_oauth: boolean;
  github_client_id: string;
  system_name: string;
  home_page_link: string;
  footer_html: string;
  wechat_qrcode: string;
  wechat_login: boolean;
  server_address: string;
  turnstile_check: boolean;
  turnstile_site_key: string;
}
