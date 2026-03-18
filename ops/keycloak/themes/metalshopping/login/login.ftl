<!DOCTYPE html>
<html class="ms-html" lang="${(locale.currentLanguageTag!'pt-BR')?html}">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>${(realm.displayName!'MetalShopping')?html}</title>
    <link rel="icon" href="${url.resourcesPath}/img/logo_ico.jpg">
    <link rel="stylesheet" href="${url.resourcesPath}/css/login.css">
</head>
<body class="ms-body">
    <div class="ms-page">
        <aside class="ms-brand-panel">
            <div class="ms-brand-content">
                <p class="ms-eyebrow">Metal Nobre Acabamentos</p>
                <h1 class="ms-title">
                    Precificacao
                    <br>
                    <span>inteligente.</span>
                </h1>
                <p class="ms-subtitle">
                    Compare precos de mercado, acompanhe tendencias e opere com
                    uma sessao web protegida pelo backend do MetalShopping.
                </p>

                <div class="ms-stats" aria-label="Destaques do MetalShopping">
                    <div class="ms-stat-card">
                        <strong>OIDC</strong>
                        <span>Login corporativo</span>
                    </div>
                    <div class="ms-stat-card">
                        <strong>RLS</strong>
                        <span>Tenant isolado</span>
                    </div>
                    <div class="ms-stat-card">
                        <strong>IAM</strong>
                        <span>Permissoes no core</span>
                    </div>
                </div>
            </div>
        </aside>

        <main class="ms-form-area">
            <div class="ms-card">
                <header class="ms-header">
                    <img class="ms-logo" src="${url.resourcesPath}/img/logo_metal_nobre.svg" alt="Metal Nobre Acabamentos">
                    <h2>Bem-vindo ao Metal<span>Shopping</span></h2>
                    <p>Entre com sua conta segura para acessar as superficies operacionais do MetalShopping.</p>
                </header>

                <section class="ms-session-panel" aria-labelledby="ms-login-title">
                    <div class="ms-copy-block">
                        <p class="ms-panel-eyebrow">Sessao protegida</p>
                        <h3 class="ms-panel-title" id="ms-login-title">Entrar com conta segura</h3>
                        <p class="ms-panel-text">
                            O navegador nao armazena token da aplicacao. A autenticacao e a
                            sessao sao controladas pelo backend com cookie HttpOnly.
                        </p>
                    </div>

                    <ul class="ms-feature-list">
                        <li>OIDC e issuer externo reais</li>
                        <li>Tenancy e autorizacao resolvidas no server_core</li>
                        <li>Mesmo modelo de identidade para web e app futuro</li>
                    </ul>

                    <#if message?has_content>
                        <div class="ms-alert ms-alert-${message.type}">
                            ${message.summary?html}
                        </div>
                    </#if>

                    <#if realm.password>
                        <form id="kc-form-login" class="ms-form" action="${url.loginAction}" method="post">
                            <#assign showUsernameField = !(usernameHidden?? && usernameHidden)>

                            <#if showUsernameField>
                                <div class="ms-field">
                                    <label class="ms-label" for="username">Email ou usuario</label>
                                    <input
                                        id="username"
                                        class="ms-input"
                                        name="username"
                                        type="text"
                                        value="${(login.username!'')?html}"
                                        autocomplete="username"
                                        autofocus
                                    >
                                </div>
                            <#else>
                                <input type="hidden" id="username" name="username" value="${(login.username!'')?html}">
                            </#if>

                            <div class="ms-field">
                                <label class="ms-label" for="password">Senha</label>
                                <input
                                    id="password"
                                    class="ms-input"
                                    name="password"
                                    type="password"
                                    autocomplete="current-password"
                                >
                            </div>

                            <div class="ms-row">
                                <#if realm.rememberMe && showUsernameField>
                                    <label class="ms-checkbox">
                                        <input id="rememberMe" name="rememberMe" type="checkbox" <#if login.rememberMe?? && login.rememberMe>checked</#if>>
                                        <span>Lembrar sessao</span>
                                    </label>
                                </#if>

                                <#if realm.resetPasswordAllowed>
                                    <a class="ms-link" href="${url.loginResetCredentialsUrl}">Esqueci minha senha</a>
                                </#if>
                            </div>

                            <button class="ms-submit" id="kc-login" name="login" type="submit">
                                Entrar com identidade segura
                            </button>
                        </form>
                    </#if>
                </section>

                <footer class="ms-footer">
                    <span>Metal<span>Shopping</span> v3.0</span>
                    <small>Auth: Keycloak + OIDC</small>
                </footer>
            </div>
        </main>
    </div>
</body>
</html>
