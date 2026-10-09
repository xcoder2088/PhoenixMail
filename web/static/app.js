const state={folder:'inbox',mails:[],selected:null,counts:{},draftId:'',timer:null,filter:'all',accounts:[],activeAccountId:'',defaultAccountId:'',searchSeq:0,brandFooter:true};
const app=document.querySelector('#app');
const icon=n=>{
  const p={
    compose:'<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M12 20h9"/><path d="M16.5 3.5a2.12 2.12 0 0 1 3 3L8 18l-4 1 1-4Z"/></svg>',
    mail:'<svg viewBox="0 0 24 24" aria-hidden="true"><rect x="3" y="5" width="18" height="14" rx="2"/><path d="m4 7 8 6 8-6"/></svg>',
    inbox:'<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M4 4h16v12H4z"/><path d="M4 13h4l2 3h4l2-3h4"/></svg>',
    star:'<svg viewBox="0 0 24 24" aria-hidden="true"><path d="m12 3 2.8 5.7 6.2.9-4.5 4.4 1.1 6.2-5.6-3-5.6 3 1.1-6.2L3 9.6l6.2-.9Z"/></svg>',
    draft:'<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M4 3h11l5 5v13H4z"/><path d="M15 3v6h5"/><path d="m8 16 5.8-5.8 2 2L10 18H8Z"/></svg>',
    sent:'<svg viewBox="0 0 24 24" aria-hidden="true"><path d="m3 11 18-8-8 18-2-7Z"/><path d="M11 14 21 3"/></svg>',
    junk:'<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M12 3 19 6v5c0 4.7-3 8.4-7 10-4-1.6-7-5.3-7-10V6z"/><path d="M12 8v4M12 15v.01"/></svg>',
    trash:'<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M4 7h16"/><path d="M9 7V4h6v3"/><path d="m7 7 1 13h8l1-13"/><path d="M10 11v5M14 11v5"/></svg>',
    chevron:'<svg viewBox="0 0 24 24" aria-hidden="true"><path d="m6 9 6 6 6-6"/></svg>',
    plus:'<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M12 5v14M5 12h14"/></svg>',
    folder:'<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M3 6h7l2 2h9v10H3z"/></svg>',
    settings:'<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M12 8a4 4 0 1 0 0 8 4 4 0 0 0 0-8Z"/><path d="m4.9 7.2 1.2-2.1 2.5 1a8 8 0 0 1 2.4-1l.4-2.7h2.4l.4 2.7a8 8 0 0 1 2.4 1l2.5-1 1.2 2.1-2.1 1.7a8 8 0 0 1 .5 2.5l2.5 1.1v2.4l-2.5 1.1a8 8 0 0 1-.5 2.5l2.1 1.7-1.2 2.1-2.5-1a8 8 0 0 1-2.4 1l-.4 2.7h-2.4l-.4-2.7a8 8 0 0 1-2.4-1l-2.5 1-1.2-2.1 2.1-1.7a8 8 0 0 1-.5-2.5L2.4 15v-2.4l2.5-1.1a8 8 0 0 1 .5-2.5Z"/></svg>',
    search:'<svg viewBox="0 0 24 24" aria-hidden="true"><circle cx="11" cy="11" r="6.5"/><path d="m16 16 4.5 4.5"/></svg>',
    filter:'<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M4 5h16l-6 7v5l-4 2v-7Z"/></svg>',
    back:'<svg viewBox="0 0 24 24" aria-hidden="true"><path d="m15 6-6 6 6 6"/></svg>',
    forward:'<svg viewBox="0 0 24 24" aria-hidden="true"><path d="m9 6 6 6-6 6"/></svg>',
    archive:'<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M4 7h16v13H4z"/><path d="M3 4h18v3H3z"/><path d="M9 12h6"/></svg>',
    tag:'<svg viewBox="0 0 24 24" aria-hidden="true"><path d="m4 5 7-2 10 10-8 8L3 11Z"/><path d="M8 8h.01"/></svg>',
    more:'<svg viewBox="0 0 24 24" aria-hidden="true"><circle cx="5" cy="12" r="1.3"/><circle cx="12" cy="12" r="1.3"/><circle cx="19" cy="12" r="1.3"/></svg>',
    paperclip:'<svg viewBox="0 0 24 24" aria-hidden="true"><path d="m8 12 6-6a4 4 0 0 1 6 6l-7 7a5 5 0 0 1-7-7l7-7"/></svg>',
    maximize:'<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M8 4H4v4M16 4h4v4M4 16v4h4M20 16v4h-4"/></svg>',
    minimize:'<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M5 12h14"/></svg>',
    close:'<svg viewBox="0 0 24 24" aria-hidden="true"><path d="m6 6 12 12M18 6 6 18"/></svg>',
    link:'<svg viewBox="0 0 24 24" aria-hidden="true"><path d="m9 15 6-6"/><path d="M7 17H6a4 4 0 0 1 0-8h4M17 7h1a4 4 0 0 1 0 8h-4"/></svg>',
    image:'<svg viewBox="0 0 24 24" aria-hidden="true"><rect x="4" y="5" width="16" height="14" rx="2"/><circle cx="9" cy="10" r="1.5"/><path d="m5 17 4-4 3 3 2-2 5 5"/></svg>',
    bullets:'<svg viewBox="0 0 24 24" aria-hidden="true"><circle cx="5" cy="6" r="1"/><circle cx="5" cy="12" r="1"/><circle cx="5" cy="18" r="1"/><path d="M9 6h10M9 12h10M9 18h10"/></svg>',
    list:'<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M8 6h12M8 12h12M8 18h12"/><circle cx="4" cy="6" r="1"/><circle cx="4" cy="12" r="1"/><circle cx="4" cy="18" r="1"/></svg>',
    spark:'<svg viewBox="0 0 24 24" aria-hidden="true"><path d="m12 3 1.5 5.5L19 10l-5.5 1.5L12 17l-1.5-5.5L5 10l5.5-1.5Z"/><path d="m19 15 .7 2.3L22 18l-2.3.7L19 21l-.7-2.3L16 18l2.3-.7Z"/></svg>',
    mic:'<svg viewBox="0 0 24 24" aria-hidden="true"><rect x="9" y="3" width="6" height="11" rx="3"/><path d="M5 11a7 7 0 0 0 14 0M12 18v3M9 21h6"/></svg>',
    globe:'<svg viewBox="0 0 24 24" aria-hidden="true"><circle cx="12" cy="12" r="8"/><path d="M4 12h16M12 4a12 12 0 0 1 0 16M12 4a12 12 0 0 0 0 16"/></svg>',
    refresh:'<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M20 7v5h-5"/><path d="M20 12a8 8 0 1 1-2.3-5.7L20 9"/></svg>'
  }; return `<span class="ui-icon">${p[n]||p.more}</span>`;
};
app.innerHTML=`
<div class="app-shell">
  <aside class="sidebar">
    <div class="brand"><img src="brand/icon-256.png"><div class="brand-copy"><strong>Phoenix<span>Mail</span></strong><small>Libre • Privé • Intelligent</small></div><button class="sidebar-toggle" id="sidebar-toggle" title="Réduire le menu" aria-label="Réduire ou ouvrir le menu">${icon('back')}</button></div>
    <button class="compose-btn" id="compose">${icon('compose')}<span>Nouveau courriel</span></button>
    <nav class="nav" id="nav">
      <button data-folder="inbox" class="active">${icon('inbox')}<label>Inbox</label><b id="c-inbox">0</b></button>
      <button data-folder="important">${icon('star')}<label>Important</label><b id="c-important">0</b></button>
      <button data-folder="drafts">${icon('draft')}<label>Brouillons</label><b id="c-drafts">0</b></button>
      <button data-folder="sent">${icon('sent')}<label>Envoyés</label><b id="c-sent">0</b></button>
      <button data-folder="archive">${icon('archive')}<label>Archives</label><b id="c-archive">0</b></button>
      <button data-folder="trash">${icon('trash')}<label>Corbeille</label><b id="c-trash">0</b></button>
    </nav>
    <section class="side-section"><div class="section-title">Comptes ${icon('chevron')}</div>
      <div id="sidebar-accounts"></div><button class="add-account" id="add-account-sidebar">${icon('plus')}<span>Ajouter un compte</span></button>
    </section>
    <section class="side-section"><div class="section-title">Dossiers ${icon('chevron')}</div>
      <div class="folder-row">${icon('folder')}<span>Personnel</span></div><div class="folder-row">${icon('folder')}<span>Travail</span></div><div class="folder-row">${icon('folder')}<span>Projets</span></div><div class="folder-row">${icon('folder')}<span>Factures</span></div><div class="folder-row">${icon('folder')}<span>Voyages</span></div><div class="folder-row">${icon('folder')}<span>Archives</span></div>
    </section>
    <button class="settings">${icon('settings')}<span>Paramètres</span></button>
  </aside>
  <section class="mail-list-panel">
    <header class="topbar"><div class="search-wrap">${icon('search')}<input id="search" placeholder="Rechercher dans les courriels..."></div><button class="refresh-mail" id="refresh-mail" title="Actualiser la boîte de réception" aria-label="Actualiser la boîte de réception">${icon('refresh')}</button><span class="sync-status" id="sync-status" role="status" aria-live="polite"></span><button class="filter" title="Filtrer les courriels" aria-label="Filtrer les courriels">${icon('filter')}</button></header>
    <div class="list-tabs"><button class="active">Tous</button><button>Non lus</button><button>Avec pièce jointe</button></div>
    <div class="list-title"><h1 id="folder-title">Inbox</h1><span id="folder-count">0</span></div>
    <div id="mail-list"></div>
  </section>
  <div class="pane-resizer" id="pane-resizer" role="separator" aria-orientation="vertical" aria-label="Redimensionner la liste des courriels" aria-valuemin="300" aria-valuemax="760" aria-valuenow="468" aria-controls="mail-list reader" tabindex="0" title="Glisser pour redimensionner · Double-cliquer pour réinitialiser"></div>
  <main class="reader" id="reader"><div class="reader-empty"><img src="brand/icon-256.png"><h2>Sélectionne un courriel</h2><p>Choisis un message pour l'afficher.</p></div></main>
</div>
<div class="composer" id="composer">
  <div class="composer-head"><strong>Nouveau message</strong><div><button id="minimize">${icon('minimize')}</button><button id="expand">${icon('maximize')}</button><button id="close">${icon('close')}</button></div></div>
  <div class="field sender-field"><label>De</label><select id="from-account" aria-label="Compte d’envoi"></select></div>
  <div class="field recipient-field"><label>À</label><input id="to" placeholder="Destinataire"><button type="button" class="recipient-toggle" id="show-cc">Cc</button><button type="button" class="recipient-toggle" id="show-bcc">Cci</button></div>
  <div class="field optional-recipient" id="cc-row" hidden><label>Cc</label><input id="cc" placeholder="Destinataires en copie"></div>
  <div class="field optional-recipient" id="bcc-row" hidden><label>Cci</label><input id="bcc" placeholder="Destinataires en copie cachée"></div>
  <div class="field subject-field"><label>Sujet</label><input id="subject" placeholder="Sujet"></div>
  <div class="formatbar">
    <select class="format-select" id="format-block" aria-label="Style de paragraphe">
      <option value="p">Paragraphe</option><option value="h2">Titre</option><option value="h3">Sous-titre</option>
    </select>
    <button type="button" class="format-btn" data-cmd="bold" title="Gras" aria-label="Gras"><strong>B</strong></button>
    <button type="button" class="format-btn" data-cmd="italic" title="Italique" aria-label="Italique"><em>I</em></button>
    <button type="button" class="format-btn" data-cmd="underline" title="Souligné" aria-label="Souligné"><u>U</u></button>
    <button type="button" class="format-btn icon-btn" data-cmd="insertUnorderedList" title="Liste à puces" aria-label="Liste à puces">${icon('bullets')}</button>
    <button type="button" class="format-btn icon-btn" data-cmd="insertOrderedList" title="Liste numérotée" aria-label="Liste numérotée">${icon('list')}</button>
    <button type="button" class="format-btn icon-btn" id="insert-link" title="Insérer un lien" aria-label="Insérer un lien">${icon('link')}</button>
    <button type="button" class="format-btn icon-btn" id="insert-image" title="Insérer une image" aria-label="Insérer une image">${icon('image')}</button>
    <button type="button" class="format-btn icon-btn" id="attach-file" title="Joindre un fichier" aria-label="Joindre un fichier">${icon('paperclip')}</button>
    <button type="button" class="ai-trigger" id="ai-trigger">${icon('spark')} IA</button>
    <input type="file" id="image-input" accept="image/*" hidden>
    <input type="file" id="attachment-input" multiple hidden>
  </div>
  <div class="editor" id="editor" contenteditable="true" spellcheck="true"></div>
  <div class="ai-bar"><button data-ai="correct">${icon('spark')} Corriger</button><button data-ai="rewrite">${icon('compose')} Reformuler</button><button data-ai="translate">${icon('globe')} Traduire</button><button data-ai="shorten">${icon('filter')} Raccourcir</button><button data-ai="reply">${icon('sent')} Réponse</button></div>
  <div class="composer-footer"><span class="composer-status" id="send-status" aria-live="polite">Prêt</span><button class="send" id="send">${icon('sent')} <span id="send-label">Envoyer</span> <small>⌄</small></button></div>
  <div class="ai-result" id="ai-result"><div class="ai-title"><span>Proposition PhoenixMail</span><button id="ai-close">×</button></div><pre id="ai-text"></pre><button id="insert">Insérer</button></div>
</div>`;

const list=document.querySelector('#mail-list'),reader=document.querySelector('#reader');
const normalizeMail=m=>({ID:String(m.id??m.ID??''),From:String(m.from??m.From??''),Email:String(m.email??m.Email??''),To:String(m.to??m.To??''),Cc:String(m.cc??m.Cc??''),Bcc:String(m.bcc??m.Bcc??''),Subject:String(m.subject??m.Subject??''),Preview:String(m.preview??m.Preview??''),Body:String(m.body??m.Body??''),Time:String(m.time??m.Time??''),Date:String(m.date??m.Date??''),Folder:String(m.folder??m.Folder??''),Read:Boolean(m.read??m.Read),Starred:Boolean(m.starred??m.Starred),Important:Boolean(m.important??m.Important),Attachments:Array.isArray(m.attachments)?m.attachments:(Array.isArray(m.Attachments)?m.Attachments:[])});
const esc=s=>String(s??'').replace(/[&<>"']/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));
function avatar(m){return `<div class="avatar">${esc((m.From||'?').split(' ').map(x=>x[0]).slice(0,2).join(''))}</div>`}
function updateCounts(){for(const k of ['inbox','important','drafts','sent','archive','trash'])document.querySelector('#c-'+k).textContent=state.counts[k]??0;document.querySelector('#folder-count').textContent=state.counts[state.folder]??0}
function render(){let items=state.mails;if(state.filter==='unread')items=items.filter(m=>!m.Read);if(state.filter==='attachment')items=items.filter(m=>m.Attachments.length>0);const searching=document.querySelector('#search').value.trim().length>0;list.innerHTML=items.length?items.map(m=>`<article class="mail ${m.Read?'':'unread'} ${state.selected===m.ID?'selected':''}" data-id="${esc(m.ID)}">${avatar(m)}<div class="mail-main"><div class="mail-row"><strong>${esc(m.From||'Inconnu')}</strong><time>${esc(m.Time)}</time></div><div class="mail-subject">${esc(m.Subject||'(Sans objet)')}</div><div class="mail-preview">${esc(m.Preview)}</div>${searching?`<div class="mail-folder">${esc({inbox:'Inbox',important:'Important',drafts:'Brouillons',sent:'Envoyés',archive:'Archives',trash:'Corbeille'}[m.Folder]||m.Folder)}</div>`:''}</div><div class="mail-icons">${m.Attachments.length?icon('paperclip'):''}<button type="button" class="mail-delete" data-action="trash" title="Mettre à la corbeille" aria-label="Mettre à la corbeille">${icon('trash')}</button>${icon('star')}</div></article>`).join(''):`<div class="empty-list"><strong>${searching?'Aucun résultat':'Aucun courriel'}</strong><span>${searching?'Aucun message ne correspond à cette recherche.':'Ce dossier ne contient aucun message correspondant.'}</span></div>`;list.querySelectorAll('.mail').forEach(x=>{x.onclick=()=>openMail(x.dataset.id);x.oncontextmenu=e=>{e.preventDefault();e.stopPropagation();state.selected=x.dataset.id;render();openMessageContextMenu(e.clientX,e.clientY,x.dataset.id)};const del=x.querySelector('.mail-delete');if(del)del.onclick=async e=>{e.preventDefault();e.stopPropagation();await historyAction('trash',x.dataset.id)}});updateCounts()}
const contextMenu=document.querySelector('#message-context-menu');
function openMessageContextMenu(x,y,id){
  const m=state.mails.find(v=>v.ID===id); if(!m||!contextMenu)return;
  contextMenu.dataset.id=id;
  const readBtn=contextMenu.querySelector('[data-context-action="read"]');
  if(readBtn)readBtn.title=m.Read?'Marquer comme non lu':'Marquer comme lu';
  const impBtn=contextMenu.querySelector('[data-context-action="important"]');
  if(impBtn){impBtn.classList.toggle('active',!!m.Important);const label=impBtn.querySelector('label');if(label)label.textContent=m.Important?'Retirer des importants':'Ajouter aux importants'}
  contextMenu.hidden=false;
  const r=contextMenu.getBoundingClientRect();
  contextMenu.style.left=Math.min(Math.max(8,x),window.innerWidth-r.width-8)+'px';
  contextMenu.style.top=Math.min(Math.max(8,y),window.innerHeight-r.height-8)+'px';
}
function closeMessageContextMenu(){if(contextMenu)contextMenu.hidden=true}
async function contextMove(folder){const id=contextMenu?.dataset.id;if(!id)return;closeMessageContextMenu();if(folder==='trash'){await historyAction('trash',id);return}const r=await fetch(`/api/mails/${encodeURIComponent(id)}/move`,{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({folder})});if(!r.ok){console.error('PhoenixMail move failed',r.status,await r.text());return}if(state.selected===id&&folder!==state.folder){state.selected=null;reader.innerHTML=`<div class="reader-empty"><img src="brand/icon-256.png"><h2>Sélectionne un courriel</h2><p>Choisis un message pour l’afficher.</p></div>`}await load()}
function composeFromMessage(mode,id){const m=state.mails.find(v=>v.ID===id);if(!m)return;closeMessageContextMenu();state.draftId='';enterCompose();const to=document.querySelector('#to'),cc=document.querySelector('#cc'),subject=document.querySelector('#subject'),editor=document.querySelector('#editor');if(mode==='reply'){to.value=m.Email||m.From||'';subject.value=(m.Subject||'').match(/^Re:/i)?m.Subject:`Re: ${m.Subject||''}`;editor.innerText=`\n\n--- Message original ---\n${m.Body||''}`}else if(mode==='reply-all'){to.value=m.Email||m.From||'';subject.value=(m.Subject||'').match(/^Re:/i)?m.Subject:`Re: ${m.Subject||''}`;editor.innerText=`\n\n--- Message original ---\n${m.Body||''}`}else{to.value='';subject.value=(m.Subject||'').match(/^Fwd:/i)?m.Subject:`Fwd: ${m.Subject||''}`;editor.innerText=`\n\n--- Message transféré ---\nDe : ${m.From||''} <${m.Email||''}>\nObjet : ${m.Subject||''}\n\n${m.Body||''}`}editor.focus();}
document.addEventListener('click',e=>{if(contextMenu&&!contextMenu.hidden&&!contextMenu.contains(e.target))closeMessageContextMenu()});
document.addEventListener('keydown',e=>{if(e.key==='Escape'&&contextMenu&&!contextMenu.hidden){closeMessageContextMenu();return}});
contextMenu?.addEventListener('click',async e=>{const move=e.target.closest('[data-move-folder]');if(move){e.preventDefault();await contextMove(move.dataset.moveFolder);return}const b=e.target.closest('[data-context-action]');if(!b)return;e.preventDefault();const action=b.dataset.contextAction;const id=contextMenu.dataset.id;const m=state.mails.find(v=>v.ID===id);if(!m)return;switch(action){case'read':closeMessageContextMenu();await historyAction(m.Read?'unread':'read',id);break;case'reply':composeFromMessage('reply',id);break;case'reply-all':composeFromMessage('reply-all',id);break;case'forward':composeFromMessage('forward',id);break;case'archive':closeMessageContextMenu();await historyAction('archive',id);break;case'trash':closeMessageContextMenu();await historyAction('trash',id);break;case'important':closeMessageContextMenu();await historyAction('important',id);break;case'print':closeMessageContextMenu();openMail(id);setTimeout(()=>window.print(),0);break;case'move-menu':e.stopPropagation();contextMenu.querySelector('.context-submenu-wrap')?.classList.toggle('open');break}});
function openMail(id){state.selected=id;const m=state.mails.find(x=>x.ID===id);if(!m)return;m.Read=true;if(window.matchMedia('(max-width: 940px)').matches)reader.classList.add('open');render();fetch(`/api/mails/${encodeURIComponent(id)}/read`,{method:'POST'});reader.innerHTML=`<div class="reader-toolbar"><button onclick="readerBack()">${icon('back')}</button><button onclick="readerForward()">${icon('forward')}</button><button onclick="historyAction('archive')">${icon('archive')}</button><button onclick="historyAction('trash')">${icon('trash')}</button><button>${icon('folder')}</button><button>${icon('tag')}</button><button>${icon('more')}</button><div></div><button onclick="historyAction('star')" class="star ${m.Starred?'on':''}">${icon('star')}</button><button>${icon('more')}</button></div><article class="message"><div class="message-head"><h1>${esc(m.Subject)}</h1><div class="sender-line">${avatar(m)}<div><strong>${esc(m.From)}</strong> &lt;${esc(m.Email)}&gt;<br><span>À : moi⌄</span></div><time>Aujourd'hui ${esc(m.Time)}</time></div></div><div class="message-body">${esc(m.Body)}</div></article>`}
window.readerBack=()=>{state.selected=null;reader.classList.remove('open');reader.innerHTML=`<div class="reader-empty"><img src="brand/icon-256.png"><h2>Sélectionne un courriel</h2><p>Choisis un message pour l’afficher.</p></div>`};window.readerForward=()=>{};window.historyAction=async(action,id)=>{const targetId=id||state.selected;if(!targetId)return false;const r=await fetch(`/api/mails/${encodeURIComponent(targetId)}/${action}`,{method:'POST'});if(!r.ok){console.error('PhoenixMail action failed',action,r.status,await r.text());return false}if(action==='trash'&&state.selected===targetId){state.selected=null;reader.innerHTML=`<div class="reader-empty"><img src="brand/icon-256.png"><h2>Sélectionne un courriel</h2><p>Choisis un message pour l’afficher.</p></div>`}await load();return true}
document.addEventListener('keydown',e=>{if(e.key!=='Delete'||e.defaultPrevented)return;const t=e.target;if(t&&((t.tagName==='INPUT')||(t.tagName==='TEXTAREA')||(t.isContentEditable)))return;if(state.selected){e.preventDefault();historyAction('trash')}})
async function load(){const q=document.querySelector('#search').value.trim();const seq=++state.searchSeq;const folder=q?'all':state.folder;const r=await fetch(`/api/mails?folder=${encodeURIComponent(folder)}&q=${encodeURIComponent(q)}&account=${encodeURIComponent(state.activeAccountId||'')}`,{cache:'no-store'});if(seq!==state.searchSeq)return;const data=await r.json();state.mails=(data.mails||[]).map(normalizeMail);state.counts=data.counts||{};document.querySelector('#folder-title').textContent=q?'Recherche':({inbox:'Inbox',important:'Important',drafts:'Brouillons',sent:'Envoyés',archive:'Archives',trash:'Corbeille'}[state.folder]||state.folder);render();if(state.selected){const m=state.mails.find(x=>x.ID===state.selected);if(m)openMail(m.ID);else{state.selected=null;reader.innerHTML=`<div class="reader-empty"><img src="brand/icon-256.png"><h2>Sélectionne un courriel</h2><p>Choisis un message pour l’afficher.</p></div>`}}}
document.querySelector('#nav').onclick=e=>{const b=e.target.closest('button[data-folder]');if(!b)return;state.folder=b.dataset.folder;state.selected=null;document.querySelector('#search').value='';document.querySelectorAll('#nav button').forEach(x=>x.classList.toggle('active',x===b));load()};
let searchTimer;document.querySelector('#search').oninput=()=>{clearTimeout(searchTimer);searchTimer=setTimeout(load,180)};
document.querySelectorAll('.list-tabs button').forEach((b,i)=>b.onclick=()=>{state.filter=['all','unread','attachment'][i];document.querySelectorAll('.list-tabs button').forEach(x=>x.classList.toggle('active',x===b));render()});
const composer=document.querySelector('#composer');
// Compact Thunderbird-like Cc/Cci controls
const ccToggle=document.querySelector('#show-cc');
const bccToggle=document.querySelector('#show-bcc');
const ccRow=document.querySelector('#cc-row');
const bccRow=document.querySelector('#bcc-row');
ccToggle.onclick=()=>{ccRow.hidden=!ccRow.hidden; if(!ccRow.hidden) document.querySelector('#cc').focus()};
bccToggle.onclick=()=>{bccRow.hidden=!bccRow.hidden; if(!bccRow.hidden) document.querySelector('#bcc').focus()};

// Message context menu (desktop mail-client behavior)
app.insertAdjacentHTML('beforeend', `
<div class="message-context-menu" id="message-context-menu" hidden role="menu" aria-label="Actions du courriel">
  <div class="context-quick-actions">
    <button data-context-action="read" title="Marquer comme lu/non lu">${icon('mail')}</button>
    <button data-context-action="reply" title="Répondre">↩</button>
    <button data-context-action="forward" title="Transférer">↪</button>
    <button data-context-action="archive" title="Archiver">${icon('archive')}</button>
    <button data-context-action="trash" title="Supprimer">${icon('trash')}</button>
  </div>
  <div class="context-separator"></div>
  <button class="context-item" data-context-action="reply"><span>↩</span><label>Répondre</label></button>
  <button class="context-item" data-context-action="reply-all"><span>↩</span><label>Répondre à tous</label></button>
  <button class="context-item" data-context-action="forward"><span>↪</span><label>Transférer</label></button>
  <div class="context-separator"></div>
  <button class="context-item" data-context-action="archive"><span>${icon('archive')}</span><label>Archiver</label></button>
  <div class="context-submenu-wrap">
    <button class="context-item" data-context-action="move-menu"><span>${icon('folder')}</span><label>Déplacer vers</label><b>›</b></button>
    <div class="context-submenu" id="context-move-menu">
      <button data-move-folder="inbox">Inbox</button>
      <button data-move-folder="important">Important</button>
      <button data-move-folder="drafts">Brouillons</button>
      <button data-move-folder="sent">Envoyés</button>
      <button data-move-folder="archive">Archives</button>
      <button data-move-folder="trash">Corbeille</button>
    </div>
  </div>
  <button class="context-item" data-context-action="important"><span>${icon('star')}</span><label>Ajouter aux importants</label></button>
  <button class="context-item danger" data-context-action="trash"><span>${icon('trash')}</span><label>Supprimer</label></button>
  <div class="context-separator"></div>
  <button class="context-item" data-context-action="print"><span>🖨</span><label>Imprimer</label></button>
</div>`);

// PhoenixMail native confirmation dialog — never expose browser "localhost says" dialogs for important actions.
app.insertAdjacentHTML('beforeend', `
<div class="pm-dialog-overlay" id="pm-dialog" hidden>
  <div class="pm-dialog" role="dialog" aria-modal="true" aria-labelledby="pm-dialog-title" aria-describedby="pm-dialog-message">
    <div class="pm-dialog-icon">${icon('trash')}</div>
    <div class="pm-dialog-content"><h3 id="pm-dialog-title">Confirmer l’action</h3><p id="pm-dialog-message"></p></div>
    <div class="pm-dialog-actions"><button type="button" class="secondary-btn" id="pm-dialog-cancel">Annuler</button><button type="button" class="primary-btn danger-btn" id="pm-dialog-confirm">Confirmer</button></div>
  </div>
</div>`);
const pmDialog=document.querySelector('#pm-dialog');
let pmDialogResolve=null;
function showConfirm(title,message,confirmLabel='Confirmer'){
  return new Promise(resolve=>{
    pmDialogResolve=resolve;
    document.querySelector('#pm-dialog-title').textContent=title;
    document.querySelector('#pm-dialog-message').textContent=message;
    document.querySelector('#pm-dialog-confirm').textContent=confirmLabel;
    pmDialog.hidden=false;
    document.body.classList.add('dialog-open');
    setTimeout(()=>document.querySelector('#pm-dialog-cancel').focus(),0);
  });
}
function closeConfirm(result){
  if(pmDialog.hidden)return;
  pmDialog.hidden=true;
  document.body.classList.remove('dialog-open');
  const resolve=pmDialogResolve; pmDialogResolve=null;
  if(resolve)resolve(result);
}
document.querySelector('#pm-dialog-cancel').onclick=()=>closeConfirm(false);
document.querySelector('#pm-dialog-confirm').onclick=()=>closeConfirm(true);
pmDialog.addEventListener('click',e=>{if(e.target===pmDialog)closeConfirm(false)});
document.addEventListener('keydown',e=>{if(pmDialog.hidden)return;if(e.key==='Escape'){e.preventDefault();closeConfirm(false)}else if(e.key==='Enter' && document.activeElement?.closest('.pm-dialog-actions')){e.preventDefault();closeConfirm(document.activeElement.id==='pm-dialog-confirm')}});

// Settings: accounts, appearance and local AI
app.insertAdjacentHTML('beforeend', `
<div class="settings-overlay" id="settings-overlay" hidden>
  <div class="settings-panel settings-panel-large" role="dialog" aria-modal="true" aria-labelledby="settings-title">
    <div class="settings-panel-head"><div><strong id="settings-title">Paramètres</strong><span>Comptes · Apparence · IA locale</span></div><button id="settings-close" aria-label="Fermer">×</button></div>
    <div class="settings-tabs">
      <button class="settings-tab active" data-settings-tab="accounts">Comptes</button>
      <button class="settings-tab" data-settings-tab="appearance">Apparence</button>
      <button class="settings-tab" data-settings-tab="ai">IA locale</button>
    </div>
    <div class="settings-content">
      <section class="settings-page active" data-settings-page="accounts">
        <div class="settings-section-head"><div><h3>Comptes courriel</h3><p>Ajoute plusieurs adresses et choisis le compte d’envoi par défaut.</p></div><button class="primary-btn" id="account-new">${icon('plus')} Ajouter un compte</button></div>
        <div id="account-list" class="account-settings-list"></div>
        <div class="account-editor" id="account-editor" hidden>
          <div class="settings-section-head"><div><h3 id="account-editor-title">Nouveau compte</h3><p>Les mots de passe restent en mémoire de session et ne sont pas enregistrés dans data.json.</p></div><button class="secondary-btn" id="account-cancel">Annuler</button></div>
          <input id="account-id" type="hidden">
          <div class="settings-form-grid">
            <label>Nom du compte<input id="account-name" placeholder="Personnel, Travail…"></label>
            <label>Adresse courriel<input id="account-email" type="email" placeholder="prenom@example.com"><button type="button" class="secondary-btn auto-detect-btn" id="account-autodetect">Détecter automatiquement</button><small class="field-help" id="account-autodetect-status">Entre ton adresse : PhoenixMail cherchera automatiquement les serveurs IMAP/SMTP.</small></label>
            <label>Nom affiché<input id="account-display" placeholder="François Bradette"></label>
            <label id="password-field">Mot de passe SMTP / IMAP<input id="account-password" type="password" autocomplete="new-password" placeholder="Non enregistré sur disque"></label>
          </div>
          <div class="oauth-connect-box" id="oauth-connect-box" hidden>
            <div><strong>Connexion sécurisée Microsoft</strong><span id="oauth-connect-status">Ce compte utilise OAuth2. Aucun mot de passe ne sera demandé.</span></div>
            <button type="button" class="primary-btn" id="oauth-connect">Connexion Microsoft</button>
          </div>
          <div class="settings-subhead">Réception (IMAP)</div>
          <div class="settings-form-grid">
            <label>Serveur IMAP<input id="imap-host" placeholder="imap.example.com"></label>
            <label>Port<input id="imap-port" type="number" min="1" max="65535" value="993"></label>
            <label>Sécurité<select id="imap-security"><option value="ssl">SSL/TLS</option><option value="starttls">STARTTLS</option><option value="plain">Aucune</option></select></label>
            <label>Utilisateur<input id="imap-user" placeholder="adresse@example.com"></label>
          </div>
          <div class="settings-subhead">Envoi (SMTP)</div>
          <div class="settings-form-grid">
            <label>Serveur SMTP<input id="smtp-host" placeholder="smtp.example.com"></label>
            <label>Port<input id="smtp-port" type="number" min="1" max="65535" value="587"></label>
            <label>Sécurité<select id="smtp-security"><option value="starttls">STARTTLS</option><option value="ssl">SSL/TLS</option><option value="plain">Aucune</option></select></label>
            <label>Utilisateur<input id="smtp-user" placeholder="adresse@example.com"></label>
          </div>
          <div class="settings-account-actions"><label class="checkbox-row"><input type="checkbox" id="account-default"> Compte d’envoi par défaut</label><div class="settings-actions"><button class="secondary-btn" id="account-test">Tester IMAP + SMTP</button><button class="primary-btn" id="account-save">Enregistrer le compte</button></div></div>
          <div class="settings-status" id="account-status">Non testé</div>
        </div>
      </section>

      <section class="settings-page" data-settings-page="appearance">
        <h3>Thème</h3><p>Thème sombre par défaut. Tu peux choisir une palette ou suivre l’apparence du système.</p>
        <div class="brand-footer-setting">
          <label class="checkbox-row"><input type="checkbox" id="brand-footer-toggle" checked> Afficher le badge « Propulsé par PhoenixMail » dans mes courriels</label>
          <p>Petit badge discret ajouté au pied des messages envoyés. Aucun suivi ni image externe.</p>
        </div>
        <div class="theme-grid">
          <button class="theme-card" data-theme-choice="dark-phoenix"><span class="theme-swatch dark phoenix"></span><strong>Phoenix Feu</strong><small>Orange / rose · défaut</small></button>
          <button class="theme-card" data-theme-choice="dark-aurora"><span class="theme-swatch dark aurora"></span><strong>Aurora</strong><small>Violet / bleu</small></button>
          <button class="theme-card" data-theme-choice="dark-ocean"><span class="theme-swatch dark ocean"></span><strong>Ocean</strong><small>Cyan / bleu</small></button>
          <button class="theme-card" data-theme-choice="dark-ruby"><span class="theme-swatch dark ruby"></span><strong>Ruby</strong><small>Rouge / magenta</small></button>
          <button class="theme-card" data-theme-choice="dark-forest"><span class="theme-swatch dark forest"></span><strong>Forest</strong><small>Émeraude / ambre</small></button>
          <button class="theme-card" data-theme-choice="dark-sunset"><span class="theme-swatch dark sunset"></span><strong>Sunset</strong><small>Ambre / rouge</small></button>
          <button class="theme-card" data-theme-choice="system"><span class="theme-swatch system"></span><strong>Système</strong><small>Suit l’apparence et la couleur d’accent du système</small></button>
        </div>
      </section>

      <section class="settings-page" data-settings-page="ai">
        <div class="settings-section-head"><div><h3>IA locale</h3><p>Tout reste local. PhoenixMail n’envoie pas le contenu des courriels à PhoenixMail.</p></div><span class="settings-status" id="ai-status">Détection…</span></div>
        <div class="ai-model-card">
          <div><strong>Modèle recommandé : Qwen3 0.6B · Q4_0</strong><p>≈ 429 Mo · GGUF · Apache-2.0. Modèle compact pour une IA locale, mais le téléchargement reste optionnel.</p></div>
          <div class="ai-model-meta"><span id="ai-model-state">Non installé</span><span id="ai-model-size"></span></div>
          <div class="settings-actions"><button class="secondary-btn" id="ai-refresh">Détecter</button><button class="secondary-btn" id="ai-download">Télécharger le modèle</button></div>
          <div class="download-progress" id="ai-download-progress" hidden><div class="download-bar"><span id="ai-download-bar"></span></div><small id="ai-download-label"></small></div>
        </div>
        <div class="settings-divider"></div>
        <div class="settings-section-head"><div><h3>Moteur local</h3><p>llama.cpp fournit l’API locale compatible OpenAI utilisée par PhoenixMail.</p></div><span class="settings-status" id="ai-runtime-status">Non détecté</span></div>
        <div class="settings-form-grid">
          <label class="full">Endpoint local<input id="ai-endpoint" placeholder="http://127.0.0.1:8080/v1/chat/completions"></label>
          <label>Modèle actif<input id="ai-model" placeholder="Qwen3-0.6B-Q4_0.gguf"></label>
          <label>Fichier modèle<input id="ai-model-path" placeholder="~/.../models/Qwen3-0.6B-Q4_0.gguf"></label>
        </div>
        <div class="settings-actions"><button class="secondary-btn" id="ai-save">Enregistrer</button><button class="primary-btn" id="ai-start">Démarrer llama.cpp</button><button class="secondary-btn" id="ai-stop">Arrêter</button></div>
        <div class="settings-note">Le modèle officiel recommandé provient de ggml-org sur Hugging Face. La licence indiquée est Apache-2.0. PhoenixMail ne l’embarque pas dans l’application : il est téléchargé séparément.</div>
      </section>
    </div>
  </div>
</div>`);

const settingsButton=document.querySelector('.settings');
const settingsOverlay=document.querySelector('#settings-overlay');
const closeSettings=()=>{settingsOverlay.hidden=true};
settingsButton.onclick=async()=>{settingsOverlay.hidden=false; showSettingsTab('accounts'); await loadAccounts(); await loadAIManager();};
document.querySelector('#settings-close').onclick=closeSettings;
settingsOverlay.addEventListener('click',e=>{if(e.target===settingsOverlay)closeSettings()});

document.querySelectorAll('.settings-tab').forEach(btn=>btn.onclick=()=>showSettingsTab(btn.dataset.settingsTab));
function showSettingsTab(name){document.querySelectorAll('.settings-tab').forEach(b=>b.classList.toggle('active',b.dataset.settingsTab===name));document.querySelectorAll('.settings-page').forEach(p=>p.classList.toggle('active',p.dataset.settingsPage===name));}

async function updateOAuthUI(acc){
  const box=document.querySelector('#oauth-connect-box'), passwordField=document.querySelector('#password-field'), btn=document.querySelector('#oauth-connect'), status=document.querySelector('#oauth-connect-status');
  const oauth=/oauth/i.test(acc?.smtp?.authentication||window.__phoenixSMTPAuth||'');
  if(!box)return;
  box.hidden=!oauth;
  if(passwordField) passwordField.style.display=oauth?'none':'';
  if(!oauth)return;
  const id=acc?.id||document.querySelector('#account-id')?.value||'';
  if(!id){ btn.disabled=true; status.textContent='Enregistre d’abord le compte; la connexion Microsoft sera ensuite disponible.'; return; }
  btn.disabled=false; status.textContent='Vérification de la connexion Microsoft…';
  try{
    const r=await fetch('/api/oauth/microsoft/status?account='+encodeURIComponent(id),{cache:'no-store'}); const d=await r.json();
    if(!d.clientConfigured){ status.textContent='Le moteur OAuth2 est prêt, mais l’identifiant d’application Microsoft de PhoenixMail doit encore être configuré.'; btn.textContent='Configurer Microsoft'; return; }
    status.textContent=d.connected?'Compte Microsoft connecté.':'Compte OAuth2 non connecté.'; btn.textContent=d.connected?'Reconnecter Microsoft':'Connexion Microsoft';
  }catch{ status.textContent='Impossible de vérifier la connexion OAuth2.'; }
}

function openAccountEditor(acc=null){
  const ed=document.querySelector('#account-editor'); ed.hidden=false;
  document.querySelector('#account-editor-title').textContent=acc?'Modifier le compte':'Nouveau compte';
  document.querySelector('#account-id').value=acc?.id||''; document.querySelector('#account-name').value=acc?.name||''; document.querySelector('#account-email').value=acc?.email||''; document.querySelector('#account-display').value=acc?.displayName||''; document.querySelector('#account-password').value='';
  document.querySelector('#imap-host').value=acc?.imap?.host||''; document.querySelector('#imap-port').value=acc?.imap?.port||993; document.querySelector('#imap-security').value=acc?.imap?.security||'ssl'; document.querySelector('#imap-user').value=acc?.imap?.username||acc?.email||'';
  window.__phoenixSMTPAuth=acc?.smtp?.authentication||''; document.querySelector('#smtp-host').value=acc?.smtp?.host||''; document.querySelector('#smtp-port').value=acc?.smtp?.port||587; document.querySelector('#smtp-security').value=acc?.smtp?.security||'starttls'; document.querySelector('#smtp-user').value=acc?.smtp?.username||acc?.email||'';
  document.querySelector('#account-default').checked=(state.defaultAccountId||'')===(acc?.id||''); document.querySelector('#account-status').textContent='Non testé';
  const st=document.querySelector('#account-autodetect-status'); if(st) st.textContent=acc?.imap?.host?'Configuration existante. Tu peux la détecter à nouveau.':'Entre ton adresse : PhoenixMail cherchera automatiquement les serveurs IMAP/SMTP.';
  updateOAuthUI(acc);
}
function closeAccountEditor(){document.querySelector('#account-editor').hidden=true}
function renderAccounts(){
  const host=document.querySelector('#account-list'); if(!host)return;
  host.innerHTML=state.accounts.length?state.accounts.map(a=>`<div class="account-card"><div class="account-card-icon">${icon('mail')}</div><div><strong>${esc(a.displayName||a.name||a.email)}</strong><span>${esc(a.email)}</span><small>${state.defaultAccountId===a.id?'Compte par défaut':'IMAP '+(a.imap?.host||'non configuré')+' · SMTP '+(a.smtp?.host||'non configuré')}</small></div><div class="account-card-actions"><button class="secondary-btn" data-edit-account="${esc(a.id)}">Modifier</button><button class="secondary-btn danger-btn" data-delete-account="${esc(a.id)}">Supprimer</button></div></div>`).join(''):`<div class="empty-list"><strong>Aucun compte configuré</strong><span>Ajoute ton premier compte courriel pour activer l’envoi et préparer la réception IMAP.</span></div>`;
  host.querySelectorAll('[data-edit-account]').forEach(b=>b.onclick=()=>openAccountEditor(state.accounts.find(a=>a.id===b.dataset.editAccount)));
  host.querySelectorAll('[data-delete-account]').forEach(b=>b.onclick=async()=>{const ok=await showConfirm('Supprimer ce compte ?','Le compte et sa configuration locale seront retirés de PhoenixMail.','Supprimer');if(!ok)return;await fetch('/api/accounts?id='+encodeURIComponent(b.dataset.deleteAccount),{method:'DELETE'});await loadAccounts()});
  document.querySelector('#from-account').innerHTML=state.accounts.map(a=>`<option value="${esc(a.id)}">${esc(a.displayName||a.email)} · ${esc(a.email)}</option>`).join('');
  if(state.activeAccountId && state.accounts.some(a=>a.id===state.activeAccountId)) document.querySelector('#from-account').value=state.activeAccountId; else if(state.defaultAccountId && state.accounts.some(a=>a.id===state.defaultAccountId)){state.activeAccountId=state.defaultAccountId;document.querySelector('#from-account').value=state.activeAccountId}
  const side=document.querySelector('#sidebar-accounts'); if(side) side.innerHTML=state.accounts.length?state.accounts.map(a=>`<button class="account-row ${state.activeAccountId===a.id?'active':''}" data-side-account="${esc(a.id)}" title="Ouvrir la boîte de réception de ${esc(a.email)}">${icon('mail')}<span>${esc(a.email)}</span></button>`).join(''):'';
  side?.querySelectorAll('[data-side-account]').forEach(b=>b.onclick=async()=>{
    state.activeAccountId=b.dataset.sideAccount;
    const from=document.querySelector('#from-account'); if(from)from.value=state.activeAccountId;
    state.folder='inbox';state.selected=null;document.querySelector('#search').value='';
    document.querySelectorAll('#nav button').forEach(x=>x.classList.toggle('active',x.dataset.folder==='inbox'));
    reader.innerHTML=`<div class="reader-empty"><img src="brand/icon-256.png"><h2>Sélectionne un courriel</h2><p>Choisis un message pour l’afficher.</p></div>`;
    syncAccountSelectionUI();await load();await refreshMailbox(state.activeAccountId);
  });
}
async function loadPhoenixConfig(){
  try{
    const r=await fetch('/api/config',{cache:'no-store'});
    const d=await r.json();
    state.brandFooter=d.brandFooter!==false;
    const toggle=document.querySelector('#brand-footer-toggle'); if(toggle)toggle.checked=state.brandFooter;
  }catch{}
}

document.querySelector('#brand-footer-toggle').addEventListener('change',async e=>{
  state.brandFooter=e.target.checked;
  try{await fetch('/api/config',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({brandFooter:state.brandFooter})})}catch{}
});

async function loadAccounts(){try{const r=await fetch('/api/accounts',{cache:'no-store'});const d=await r.json();state.accounts=d.accounts||[];state.defaultAccountId=d.defaultAccount||'';renderAccounts();}catch{state.accounts=[];state.defaultAccountId='';renderAccounts();}}
async function autodetectAccount(){
  const email=document.querySelector('#account-email').value.trim();
  const status=document.querySelector('#account-autodetect-status');
  if(!email || !email.includes('@')){ status.textContent='Entre une adresse courriel valide.'; return; }
  status.textContent='Recherche automatique des paramètres…';
  try{
    const r=await fetch('/api/autoconfig',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({email})});
    const d=await r.json().catch(()=>({}));
    if(!r.ok){ status.textContent=d.error||'Configuration automatique introuvable. La configuration manuelle reste disponible.'; return; }
    const setIfEmpty=(sel,val)=>{const el=document.querySelector(sel); if(el && !el.value && val) el.value=val;};
    setIfEmpty('#account-display',d.shortName||d.provider||'');
    setIfEmpty('#account-name',d.shortName||d.provider||email.split('@')[1]);
    document.querySelector('#imap-host').value=d.imap?.host||'';
    document.querySelector('#imap-port').value=d.imap?.port||993;
    document.querySelector('#imap-security').value=d.imap?.security||'ssl';
    document.querySelector('#imap-user').value=d.imap?.username||email;
    document.querySelector('#smtp-host').value=d.smtp?.host||'';
    document.querySelector('#smtp-port').value=d.smtp?.port||587;
    document.querySelector('#smtp-security').value=d.smtp?.security||'starttls';
    document.querySelector('#smtp-user').value=d.smtp?.username||email;
    const auth=[d.imap?.authentication,d.smtp?.authentication].filter(Boolean);
    const oauth=auth.some(x=>/oauth/i.test(x));
    window.__phoenixSMTPAuth=d.smtp?.authentication||d.imap?.authentication||''; status.textContent=`Configuration trouvée${d.provider?` pour ${d.provider}`:''}. ${oauth?'Authentification OAuth détectée; une connexion fournisseur sera requise.':'IMAP/SMTP prêts.'}`; updateOAuthUI({id:document.querySelector('#account-id').value.trim(),smtp:{authentication:window.__phoenixSMTPAuth}});
  }catch{ status.textContent='Impossible de rechercher les paramètres. Vérifie la connexion Internet ou passe en configuration manuelle.'; }
}
document.querySelector('#account-autodetect').onclick=autodetectAccount;
document.querySelector('#account-email').addEventListener('blur',()=>{ if(document.querySelector('#account-email').value.includes('@')) autodetectAccount(); });
document.querySelector('#account-new').onclick=()=>openAccountEditor();
document.querySelector('#add-account-sidebar').onclick=async()=>{settingsOverlay.hidden=false;showSettingsTab('accounts');await loadAccounts();openAccountEditor()};
document.querySelector('#account-cancel').onclick=closeAccountEditor;
document.querySelector('#account-save').onclick=async()=>{
  const id=document.querySelector('#account-id').value.trim(); const email=document.querySelector('#account-email').value.trim(); if(!email){alert('Adresse courriel requise.');return}
  const payload={id,name:document.querySelector('#account-name').value.trim(),email,displayName:document.querySelector('#account-display').value.trim(),setDefault:document.querySelector('#account-default').checked,imap:{host:document.querySelector('#imap-host').value.trim(),port:Number(document.querySelector('#imap-port').value||993),security:document.querySelector('#imap-security').value,username:document.querySelector('#imap-user').value.trim()},smtp:{host:document.querySelector('#smtp-host').value.trim(),port:Number(document.querySelector('#smtp-port').value||587),security:document.querySelector('#smtp-security').value,authentication:(window.__phoenixSMTPAuth||''),username:document.querySelector('#smtp-user').value.trim(),from:email,displayName:document.querySelector('#account-display').value.trim(),password:document.querySelector('#account-password').value}};
  const r=await fetch('/api/accounts',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(payload)});const d=await r.json().catch(()=>({}));if(!r.ok){alert(d.error||'Impossible d’enregistrer le compte.');return}closeAccountEditor();await loadAccounts();
};
document.querySelector('#oauth-connect').onclick=async()=>{
  const id=document.querySelector('#account-id').value.trim();
  if(!id){document.querySelector('#oauth-connect-status').textContent='Enregistre d’abord le compte.';return;}
  const btn=document.querySelector('#oauth-connect'); btn.disabled=true; document.querySelector('#oauth-connect-status').textContent='Ouverture de la connexion Microsoft…';
  try{
    const r=await fetch('/oauth/microsoft/start?account='+encodeURIComponent(id)); const d=await r.json().catch(()=>({}));
    if(!r.ok){document.querySelector('#oauth-connect-status').textContent=d.error||'Impossible de démarrer OAuth2 Microsoft.'; btn.disabled=false; return;}
    document.querySelector('#oauth-connect-status').textContent='Navigateur ouvert. Termine la connexion Microsoft puis reviens ici.';
    setTimeout(()=>updateOAuthUI(state.accounts.find(a=>a.id===id)),2500);
  }catch{document.querySelector('#oauth-connect-status').textContent='Impossible de contacter PhoenixMail.'; btn.disabled=false;}
};
document.querySelector('#account-test').onclick=async()=>{const id=document.querySelector('#account-id').value.trim();if(!id){alert('Enregistre d’abord le compte pour le tester.');return}const payload={accountId:id,imap:true,smtp:true,password:document.querySelector('#account-password').value,imapPassword:document.querySelector('#account-password').value};const r=await fetch('/api/accounts/test',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(payload)});const d=await r.json().catch(()=>({}));document.querySelector('#account-status').textContent=[d.smtp,d.imap].filter(Boolean).join(' · ')||(d.error||'Échec du test')};
// The composer send-account selector is deliberately independent from the mailbox currently being viewed.

async function loadAIManager(){
  try{const r=await fetch('/api/ai/models',{cache:'no-store'});const d=await r.json(); const rec=d.recommended; const found=(d.models||[]).find(m=>m.name===rec.name); document.querySelector('#ai-model-state').textContent=found?'Installé':'Non installé'; document.querySelector('#ai-model-size').textContent=found?formatBytes(found.size):'≈ 429 Mo'; const cfg=d.ai||{};document.querySelector('#ai-endpoint').value=cfg.endpoint||'http://127.0.0.1:8080/v1/chat/completions';document.querySelector('#ai-model').value=cfg.model||'';document.querySelector('#ai-model-path').value=cfg.modelPath||found?.path||'';const dl=d.download||{}; if(dl.state==='downloading'){document.querySelector('#ai-download-progress').hidden=false;const pct=dl.total>0?Math.round(dl.bytes*100/dl.total):0;document.querySelector('#ai-download-bar').style.width=pct+'%';document.querySelector('#ai-download-label').textContent=`Téléchargement ${pct}% · ${formatBytes(dl.bytes)} / ${formatBytes(dl.total)}`;}else{document.querySelector('#ai-download-progress').hidden=true;} document.querySelector('#ai-runtime-status').textContent=(dl.state==='ready'||found)?'Modèle disponible':'Modèle non installé'; document.querySelector('#ai-status').textContent=cfg.modelPath?'Modèle configuré':'Moteur local non configuré';}
  catch{document.querySelector('#ai-status').textContent='Indisponible'}
}
function formatBytes(n){if(!n||n<0)return'';if(n<1024)return`${n} o`;if(n<1024*1024)return`${(n/1024).toFixed(0)} Ko`;return`${(n/1024/1024).toFixed(0)} Mo`}
document.querySelector('#ai-download').onclick=async()=>{const r=await fetch('/api/ai/download',{method:'POST'});if(!r.ok){const d=await r.json().catch(()=>({}));alert(d.error||'Téléchargement impossible.');return}loadAIManager()};
document.querySelector('#ai-refresh').onclick=loadAIManager;
document.querySelector('#ai-save').onclick=async()=>{const payload={ai:{endpoint:document.querySelector('#ai-endpoint').value.trim(),model:document.querySelector('#ai-model').value.trim(),modelPath:document.querySelector('#ai-model-path').value.trim(),autoStart:false}};const r=await fetch('/api/config',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(payload)});if(!r.ok){alert('Impossible d’enregistrer la configuration IA.');return}await loadAIManager()};
document.querySelector('#ai-start').onclick=async()=>{const r=await fetch('/api/ai/start',{method:'POST'});const d=await r.json().catch(()=>({}));if(!r.ok)alert(d.error||'Impossible de démarrer llama.cpp. Installez llama-server puis réessayez.');else await loadAIManager()};
document.querySelector('#ai-stop').onclick=async()=>{const r=await fetch('/api/ai/stop',{method:'POST'});if(r.ok)await loadAIManager()};

const systemScheme=window.matchMedia?.('(prefers-color-scheme: dark)');
const systemAccentMap={
  blue:['#5b7cff','#4facfe'], teal:['#16b8b3','#3cc6c0'], green:['#39bf63','#71d48c'], yellow:['#d8a52d','#f2c94c'],
  orange:['#ff8a1f','#ffb04a'], red:['#ef5361','#ff7d87'], pink:['#ef4f9a','#ff79ba'], purple:['#8a66ff','#b18cff'], slate:['#738196','#9eabbc']
};
async function fetchSystemAppearance(){
  try{const r=await fetch('/api/system-appearance',{cache:'no-store'}); if(!r.ok) return null; return await r.json();}
  catch{return null}
}
async function applySystemTheme(persist=false){
  // Establish the system theme immediately with a non-Phoenix fallback so we never flash orange.
  const root=document.documentElement;
  root.dataset.theme='system';
  root.dataset.themeChoice='system';
  root.style.setProperty('--accent','#5b7cff');
  root.style.setProperty('--accent2','#4facfe');
  root.style.setProperty('--icon-accent','#4facfe');
  const appearance=await fetchSystemAppearance();
  const light=appearance?.scheme==='light' || (!appearance?.scheme && systemScheme?.matches===false);
  const [accent,accent2]=systemAccentMap[appearance?.accent]||systemAccentMap.blue;
  root.dataset.theme=light?'system-light':'system';
  root.dataset.themeChoice='system';
  root.style.setProperty('--accent',accent);
  root.style.setProperty('--accent2',accent2);
  root.style.setProperty('--icon-accent',accent2);
  root.style.setProperty('--accent-soft', light ? `color-mix(in srgb, ${accent} 15%, #ffffff)` : `color-mix(in srgb, ${accent} 18%, #1a1e29)`);
  root.style.setProperty('--badge-bg', light ? `color-mix(in srgb, ${accent} 10%, #e2e6f0)` : `color-mix(in srgb, ${accent} 15%, #2a3044)`);
  root.style.setProperty('--selected-bg', light ? `color-mix(in srgb, ${accent} 15%, #f5f7fb)` : `color-mix(in srgb, ${accent} 16%, #1a1e29)`);
  root.style.setProperty('--selected-bg2', light ? '#f6f7fb' : '#1b2232');
  root.style.setProperty('--purple',accent);
  root.style.setProperty('--line', light ? '#d6dbe6' : '#2b303c');
  root.style.setProperty('--muted', light ? '#687289' : '#9ba2b1');
  root.style.setProperty('--text', light ? '#20263a' : '#f5f6fa');
  if(persist) localStorage.setItem('phoenixmail.theme','system');
  document.querySelectorAll('[data-theme-choice]').forEach(b=>b.classList.toggle('selected',b.dataset.themeChoice==='system'));
}
function applyStaticTheme(choice,persist=true){
  document.documentElement.dataset.theme=choice;
  document.documentElement.dataset.themeChoice=choice;
  document.documentElement.style.cssText='';
  if(persist) localStorage.setItem('phoenixmail.theme',choice);
  document.querySelectorAll('[data-theme-choice]').forEach(b=>b.classList.toggle('selected',b.dataset.themeChoice===choice));
}
function applyTheme(choice,persist=true){
  if(choice==='system') return applySystemTheme(persist);
  applyStaticTheme(choice,persist);
}
document.querySelectorAll('[data-theme-choice]').forEach(b=>b.onclick=()=>applyTheme(b.dataset.themeChoice));
const savedTheme=localStorage.getItem('phoenixmail.theme')||'dark-phoenix';
applyTheme(savedTheme,false);
if(systemScheme) systemScheme.addEventListener?.('change',()=>{if(document.documentElement.dataset.themeChoice==='system') applySystemTheme(false)});
const sidebar=document.querySelector('.sidebar');
const mainShell=document.querySelector('.app-shell');
const sidebarToggle=document.querySelector('#sidebar-toggle');
function setSidebar(mode){
  const collapsed=mode==='collapsed';
  mainShell.classList.toggle('sidebar-collapsed', collapsed);
  mainShell.classList.remove('sidebar-hidden');
  localStorage.setItem('phoenixmail.sidebarMode', collapsed?'collapsed':'expanded');
  sidebarToggle.innerHTML=icon(collapsed?'forward':'back');
  sidebarToggle.title=collapsed?'Ouvrir le menu':'Réduire le menu';
  sidebarToggle.setAttribute('aria-label', collapsed?'Ouvrir le menu':'Réduire le menu');
}
sidebarToggle.onclick=()=>{
  const collapsed=mainShell.classList.contains('sidebar-collapsed');
  setSidebar(collapsed?'expanded':'collapsed');
};
setSidebar(localStorage.getItem('phoenixmail.sidebarMode')||'expanded');
// Resizable message-list / reader split. The CSS grid consumes --list-width on this shell.
const paneResizer=document.querySelector('#pane-resizer');
const listPanel=document.querySelector('.mail-list-panel');
const savedListWidth=Number(localStorage.getItem('phoenixmail.listWidth')||468);
function listWidthBounds(){
  const shellWidth=mainShell.getBoundingClientRect().width;
  const sidebarWidth=sidebar.getBoundingClientRect().width;
  const dividerWidth=8;
  const minReaderWidth=320;
  const max=Math.max(300,Math.min(760,Math.floor(shellWidth-sidebarWidth-dividerWidth-minReaderWidth)));
  return {min:Math.min(300,max),max};
}
function clampListWidth(v){const b=listWidthBounds();return Math.max(b.min,Math.min(b.max,Math.round(Number(v)||468)));}
function setListWidth(v,persist=true){
  const w=clampListWidth(v);
  mainShell.style.setProperty('--list-width',w+'px');
  paneResizer.setAttribute('aria-valuemin',String(listWidthBounds().min));
  paneResizer.setAttribute('aria-valuemax',String(listWidthBounds().max));
  paneResizer.setAttribute('aria-valuenow',String(w));
  if(persist)localStorage.setItem('phoenixmail.listWidth',String(w));
  return w;
}
setListWidth(savedListWidth,false);
let resizeState=null;
paneResizer.addEventListener('pointerdown',e=>{
  if(window.matchMedia('(max-width: 940px)').matches)return;
  resizeState={startX:e.clientX,startWidth:listPanel.getBoundingClientRect().width,pointerId:e.pointerId};
  paneResizer.setPointerCapture?.(e.pointerId);
  document.body.classList.add('resizing-panes');
  e.preventDefault();
});
paneResizer.addEventListener('pointermove',e=>{if(!resizeState)return;setListWidth(resizeState.startWidth+(e.clientX-resizeState.startX),false)});
const finishResize=e=>{
  if(!resizeState)return;
  setListWidth(listPanel.getBoundingClientRect().width,true);
  const pointerId=resizeState.pointerId;
  resizeState=null;
  document.body.classList.remove('resizing-panes');
  try{if(pointerId!=null&&paneResizer.hasPointerCapture?.(pointerId))paneResizer.releasePointerCapture(pointerId)}catch{}
};
paneResizer.addEventListener('pointerup',finishResize);
paneResizer.addEventListener('pointercancel',finishResize);
paneResizer.addEventListener('lostpointercapture',finishResize);
paneResizer.addEventListener('dblclick',()=>setListWidth(468));
paneResizer.addEventListener('keydown',e=>{
  const current=listPanel.getBoundingClientRect().width;
  const step=e.shiftKey?40:16;
  if(e.key==='ArrowLeft'){e.preventDefault();setListWidth(current-step,true)}
  else if(e.key==='ArrowRight'){e.preventDefault();setListWidth(current+step,true)}
  else if(e.key==='Home'){e.preventDefault();setListWidth(300,true)}
  else if(e.key==='End'){e.preventDefault();setListWidth(760,true)}
});
window.addEventListener('resize',()=>{
  if(resizeState)finishResize();
  const preferred=Number(localStorage.getItem('phoenixmail.listWidth')||468);
  setListWidth(preferred,false);
});
function resetComposeFields(){
  clearTimeout(state.timer);
  state.timer=null;
  state.draftId='';
  document.querySelector('#to').value='';
  document.querySelector('#cc').value='';
  document.querySelector('#bcc').value='';
  document.querySelector('#subject').value='';
  document.querySelector('#editor').innerText='';
  document.querySelector('#cc-row').hidden=true;
  document.querySelector('#bcc-row').hidden=true;
  document.querySelector('#ai-result').classList.remove('open');
  document.querySelector('#send').disabled=false;
  document.querySelector('#send').classList.remove('sending');
  document.querySelector('#send-label').textContent='Envoyer';
  setSendStatus('idle','Prêt');
}
function enterCompose(){
  // A new compose window must always start clean. A previously sent message
  // must never leave its success state or text in the next compose.
  resetComposeFields();
  composer.classList.add('open','fullscreen');
  composer.classList.remove('minimized','expanded');
  document.body.classList.add('compose-active');
  document.querySelector('#editor').focus();
}
async function discardComposeDraft(){
  clearTimeout(state.timer);
  state.timer=null;
  const id=state.draftId;
  state.draftId='';
  if(id){
    try{await fetch(`/api/drafts/${encodeURIComponent(id)}`,{method:'DELETE'});}catch{}
  }
  resetComposeFields();
}
async function exitCompose(){
  const hasContent=Boolean(
    document.querySelector('#to').value.trim() ||
    document.querySelector('#cc').value.trim() ||
    document.querySelector('#bcc').value.trim() ||
    document.querySelector('#subject').value.trim() ||
    document.querySelector('#editor').innerText.trim()
  );
  if(hasContent || state.draftId){
    const discard=await showConfirm('Supprimer le brouillon ?','Ce brouillon sera supprimé. Cette action ne peut pas être annulée.','Supprimer');
    if(!discard)return;
    await discardComposeDraft();
  }else{
    resetComposeFields();
  }
  composer.classList.remove('open','fullscreen','expanded','minimized');
  document.body.classList.remove('compose-active');
}
document.querySelector('#compose').onclick=enterCompose;
document.querySelector('#close').onclick=()=>{exitCompose()};
document.querySelector('#minimize').onclick=()=>document.querySelector('#composer').classList.toggle('minimized');
document.querySelector('#expand').onclick=()=>composer.classList.toggle('expanded');

// Rich-text toolbar: real editing actions, visually consistent with PhoenixMail.
const editor=document.querySelector('#editor');
let savedRange=null;
const rememberSelection=()=>{const sel=window.getSelection(); if(sel&&sel.rangeCount) savedRange=sel.getRangeAt(0).cloneRange();};
editor.addEventListener('keyup',rememberSelection);
editor.addEventListener('mouseup',rememberSelection);
editor.addEventListener('focus',rememberSelection);
const restoreSelection=()=>{if(savedRange){const sel=window.getSelection();sel.removeAllRanges();sel.addRange(savedRange);}};
const runCommand=(cmd,value=null)=>{restoreSelection();editor.focus();document.execCommand(cmd,false,value);rememberSelection();};
document.querySelectorAll('.format-btn[data-cmd]').forEach(btn=>btn.addEventListener('click',()=>runCommand(btn.dataset.cmd)));
document.querySelector('#format-block').addEventListener('change',e=>runCommand('formatBlock',e.target.value));
document.querySelector('#insert-link').addEventListener('click',()=>{const url=window.prompt('URL du lien :','https://'); if(url) runCommand('createLink',url)});
const imageInput=document.querySelector('#image-input');
document.querySelector('#insert-image').addEventListener('click',()=>imageInput.click());
imageInput.addEventListener('change',()=>{const file=imageInput.files?.[0]; if(!file)return; const reader=new FileReader(); reader.onload=()=>{restoreSelection();editor.focus();document.execCommand('insertImage',false,reader.result);rememberSelection()}; reader.readAsDataURL(file); imageInput.value=''});
const attachmentInput=document.querySelector('#attachment-input');
document.querySelector('#attach-file').addEventListener('click',()=>attachmentInput.click());
attachmentInput.addEventListener('change',()=>{const files=[...(attachmentInput.files||[])]; if(files.length) alert(files.map(f=>f.name).join('\n')); attachmentInput.value=''});
document.querySelector('#ai-trigger').addEventListener('click',()=>document.querySelector('[data-ai=\"correct\"]').click());
let lastResult='';document.querySelectorAll('[data-ai]').forEach(btn=>btn.onclick=async()=>{
  const text=document.querySelector('#editor').innerText.trim();
  if(!text)return;
  try{
    const r=await fetch('/api/ai',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({action:btn.dataset.ai,text,language:'auto'})});
    const d=await r.json();
    lastResult=d.result||'';
    if(d.available===false){
      document.querySelector('#ai-text').textContent='Moteur IA local non installé dans cette version. La fonction '+btn.dataset.ai+' sera activée avec le vrai moteur local en V0.4.';
      document.querySelector('#insert').disabled=true;
    }else{
      document.querySelector('#ai-text').textContent=lastResult;
      document.querySelector('#insert').disabled=false;
    }
    document.querySelector('#ai-result').classList.add('open');
  }catch(err){
    document.querySelector('#ai-text').textContent='Le moteur IA local est indisponible.';
    document.querySelector('#insert').disabled=true;
    document.querySelector('#ai-result').classList.add('open');
  }
});
document.querySelector('#ai-close').onclick=()=>document.querySelector('#ai-result').classList.remove('open');document.querySelector('#insert').onclick=()=>{document.querySelector('#editor').innerText=lastResult;document.querySelector('#ai-result').classList.remove('open');saveDraft()};
async function saveDraft(){const body=document.querySelector('#editor').innerText;const r=await fetch('/api/drafts',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({id:state.draftId,to:document.querySelector('#to').value,cc:document.querySelector('#cc').value,bcc:document.querySelector('#bcc').value,subject:document.querySelector('#subject').value,body,accountID:document.querySelector('#from-account')?.value||state.activeAccountId})});if(r.ok){const d=await r.json();state.draftId=String(d.id??d.ID??'')}}
['#editor','#to','#cc','#bcc','#subject'].forEach(sel=>document.querySelector(sel).addEventListener('input',()=>{clearTimeout(state.timer);state.timer=setTimeout(saveDraft,500)}));
const sendButton=document.querySelector('#send'),sendLabel=document.querySelector('#send-label'),sendStatus=document.querySelector('#send-status');
function setSendStatus(kind,text){sendStatus.textContent=text;sendStatus.dataset.state=kind||'idle';}
function setSending(active){sendButton.disabled=active;sendButton.classList.toggle('sending',active);sendLabel.textContent=active?'Envoi…':'Envoyer';}
sendButton.onclick=async()=>{
  const to=document.querySelector('#to').value.trim();
  if(!to){setSendStatus('error','Destinataire requis');document.querySelector('#to').focus();return}
  const accountID=document.querySelector('#from-account')?.value||state.activeAccountId;
  const acc=state.accounts.find(a=>a.id===accountID);
  if(acc?.smtp?.authentication?.toLowerCase()==='oauth2'){
    // Do not rely on the account configuration alone: OAuth2 tokens live in the
    // local runtime and are intentionally not persisted in data.json. Ask the
    // backend for the live Microsoft OAuth status before blocking the send.
    try {
      const sr=await fetch('/api/oauth/microsoft/status?account='+encodeURIComponent(accountID),{cache:'no-store'});
      const sd=await sr.json().catch(()=>({}));
      if(!sd.connected){
        setSendStatus('error','Ce compte nécessite une connexion OAuth2. Connecte Microsoft dans Paramètres → Comptes.');
        return;
      }
    } catch {
      setSendStatus('error','Impossible de vérifier la connexion OAuth2 Microsoft.');
      return;
    }
  }
  setSending(true);
  setSendStatus('sending','Envoi en cours…');
  const controller=new AbortController();
  const timeout=setTimeout(()=>controller.abort(),45000);
  try{
    const r=await fetch('/api/send',{method:'POST',headers:{'Content-Type':'application/json'},signal:controller.signal,body:JSON.stringify({to,cc:document.querySelector('#cc').value,bcc:document.querySelector('#bcc').value,subject:document.querySelector('#subject').value,body:document.querySelector('#editor').innerText,draftID:state.draftId,accountID})});
    const d=await r.json().catch(()=>({}));
    if(r.ok){
      setSendStatus('success','Courriel envoyé');
      state.draftId='';document.querySelector('#to').value='';document.querySelector('#cc').value='';document.querySelector('#bcc').value='';document.querySelector('#subject').value='';document.querySelector('#editor').innerText='';
      setTimeout(()=>{exitCompose();state.folder='sent';document.querySelectorAll('#nav button').forEach(x=>x.classList.toggle('active',x.dataset.folder==='sent'));load()},350);
    }else{
      const msg=d.userMessage||d.error||'Impossible d’envoyer le courriel.';
      setSendStatus('error',msg);
    }
  }catch(err){
    setSendStatus('error',err?.name==='AbortError'?'Le serveur met trop de temps à répondre. Vérifie la connexion SMTP dans Paramètres.':'Impossible de joindre PhoenixMail.');
  }finally{
    clearTimeout(timeout);setSending(false);
  }
};
function syncAccountSelectionUI(){
  document.querySelectorAll('[data-side-account]').forEach(b=>{
    const active=b.dataset.sideAccount===state.activeAccountId;
    b.classList.toggle('active',active);
    if(active)b.setAttribute('aria-current','page');else b.removeAttribute('aria-current');
  });
}
async function refreshMailbox(accountID=state.activeAccountId){
  const button=document.querySelector('#refresh-mail'),status=document.querySelector('#sync-status');
  if(!accountID){status.textContent='Choisis un compte';status.title='Sélectionne un compte courriel avant l’actualisation.';return false}
  button.disabled=true;button.classList.add('busy');status.dataset.state='busy';status.textContent='Actualisation…';status.title='Connexion au serveur IMAP en cours.';
  let ok=false;
  try{
    const r=await fetch(`/api/accounts/refresh?id=${encodeURIComponent(accountID)}`,{method:'POST',cache:'no-store'});
    const d=await r.json().catch(()=>({}));
    if(!r.ok){status.dataset.state='error';status.textContent='Échec IMAP';status.title=d.error||'Actualisation impossible.';console.error('PhoenixMail IMAP refresh failed:',d.error||r.status);}
    else{ok=true;status.dataset.state='success';status.textContent=d.newMessages?`+${d.newMessages} nouveau(x)`:'Aucun nouveau';status.title=`${d.account} actualisé : ${d.newMessages} nouveau(x), ${d.fetched} message(s) récupéré(s) depuis le serveur.`;}
  }catch(err){status.dataset.state='error';status.textContent='Erreur réseau';status.title=err?.message||'Impossible de joindre PhoenixMail.';console.error('PhoenixMail refresh error:',err)}
  finally{button.disabled=false;button.classList.remove('busy');if(status.dataset.state==='busy')status.dataset.state='';await load();}
  return ok;
}
document.querySelector('#refresh-mail').onclick=()=>refreshMailbox();
loadPhoenixConfig();
loadAccounts().then(async()=>{
  if(!state.activeAccountId)state.activeAccountId=state.defaultAccountId||state.accounts[0]?.id||'';
  renderAccounts();syncAccountSelectionUI();await load();
});
