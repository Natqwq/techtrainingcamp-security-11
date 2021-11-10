var fingerprint;
function myInit() {
    if($.cookie("DeviceID") == null){
        fingerprint =  new Fingerprint().get();
        // console.log("第一次登录！")
        $.cookie("DeviceID", fingerprint, {
            expires : 365,
            path: '/'
        });
    }else{
        fingerprint = $.cookie("DeviceID");
    }
}