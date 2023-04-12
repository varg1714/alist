import{d as _,u as C,a as v,b6 as u,e as x,c8 as p,f as e,a0 as h,B as o,Y as T,c9 as B,ca as F,cb as f,v as m,cc as b,cd as P,ce as s,bd as g,n as $,W as R,bx as W,ao as D,aq as H,a2 as I}from"./index.0596c76d.js";import{o as L}from"./index.bf378032.js";import{D as M}from"./DeletePopover.a1846e76.js";import{W as q}from"./Wether.6c48c1cc.js";const z=n=>{const t=[{name:"general",color:"info"},{name:"guest",color:"neutral"},{name:"admin",color:"accent"}];return e(W,{get colorScheme(){return t[n.role].color},get children(){return t[n.role].name}})},O=n=>{const t=_(),i=a=>`$${a?"success":"danger"}9`;return e(h,{spacing:"$0_5",get children(){return e(m,{each:D,children:(a,d)=>e(H,{get label(){return t(`users.permissions.${a}`)},get children(){return e(T,{boxSize:"$2",rounded:"$full",get bg(){return i(I.can(n.user,d()))}})}})})}})},A=()=>{const n=_();L("manage.sidemenu.users");const{to:t}=C(),[i,a]=v(()=>u.get("/admin/user/list")),[d,k]=x([]),l=async()=>{const r=await a();g(r,c=>k(c.content))};l();const[S,U]=p(r=>u.post(`/admin/user/delete?id=${r}`)),[w,y]=p(r=>u.post(`/admin/user/cancel_2fa?id=${r}`));return e(R,{spacing:"$2",alignItems:"start",w:"$full",get children(){return[e(h,{spacing:"$2",get children(){return[e(o,{colorScheme:"accent",get loading(){return i()},onClick:l,get children(){return n("global.refresh")}}),e(o,{onClick:()=>{t("/@manage/users/add")},get children(){return n("global.add")}})]}}),e(T,{w:"$full",overflowX:"auto",get children(){return e(B,{highlightOnHover:!0,dense:!0,get children(){return[e(F,{get children(){return e(f,{get children(){return[e(m,{each:["username","base_path","role","permission","available"],children:r=>e(b,{get children(){return n(`users.${r}`)}})}),e(b,{get children(){return n("global.operations")}})]}})}}),e(P,{get children(){return e(m,{get each(){return d()},children:r=>e(f,{get children(){return[e(s,{get children(){return r.username}}),e(s,{get children(){return r.base_path}}),e(s,{get children(){return e(z,{get role(){return r.role}})}}),e(s,{get children(){return e(O,{user:r})}}),e(s,{get children(){return e(q,{get yes(){return!r.disabled}})}}),e(s,{get children(){return e(h,{spacing:"$2",get children(){return[e(o,{onClick:()=>{t(`/@manage/users/edit/${r.id}`)},get children(){return n("global.edit")}}),e(M,{get name(){return r.username},get loading(){return S()===r.id},onClick:async()=>{const c=await U(r.id);g(c,()=>{$.success(n("global.delete_success")),l()})}}),e(o,{colorScheme:"accent",get loading(){return w()===r.id},onClick:async()=>{const c=await y(r.id);g(c,()=>{$.success(n("users.cancel_2fa_success")),l()})},get children(){return n("users.cancel_2fa")}})]}})}})]}})})}})]}})}})]}})};export{A as default};
