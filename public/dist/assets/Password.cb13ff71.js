import{d as i,u,f as e,bf as c,I as g,_ as p,J as d,bO as m,a0 as o,bg as h,bc as a,B as s,W as f}from"./index.86895a3d.js";import{a as $}from"./Layout.c6971c82.js";import{L as b}from"./index.bd520d27.js";import"./index.29bdb5b4.js";import"./Markdown.d3253182.js";import"./api.0999634a.js";import"./useUtil.82a9a0ca.js";import"./index.4c2446cc.js";import"./FolderTree.844c3ed2.js";const B=()=>{const t=i(),{refresh:n}=$(),{back:l}=u();return e(f,{w:{"@initial":"$full","@md":"$lg"},p:"$8",spacing:"$3",alignItems:"start",get children(){return[e(c,{get children(){return t("home.input_password")}}),e(g,{type:"password",get value(){return p()},get background(){return d("$neutral3","$neutral2")()},onKeyDown:r=>{r.key==="Enter"&&n(!0)},onInput:r=>m(r.currentTarget.value)}),e(o,{w:"$full",justifyContent:"space-between",get children(){return[e(h,{fontSize:"$sm",color:"$neutral10",direction:{"@initial":"column","@sm":"row"},columnGap:"$1",get children(){return[e(a,{get children(){return t("global.have_account")}}),e(a,{color:"$info9",as:b,get href(){return`/@login?redirect=${encodeURIComponent(location.pathname)}`},get children(){return t("global.go_login")}})]}}),e(o,{spacing:"$2",get children(){return[e(s,{colorScheme:"neutral",onClick:l,get children(){return t("global.back")}}),e(s,{onClick:()=>n(!0),get children(){return t("global.ok")}})]}})]}})]}})};export{B as default};
