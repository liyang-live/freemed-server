// app.js

$(document).ready(function() {
	console.log('document.ready');

	$.ajaxSetup({
		// Set a reasonable timeout for all queries to deal with failures
		timeout: 10000,
		// Disable request caching
		cache:   false
	});

	// Force toaster notifications to show how long until they go away
	toastr.options.progressBar = true;

	if (sessionId == null) {
		var sessionIdRaw = $.sessionStorage.getItem('sessionId');
		if (sessionIdRaw != null && sessionIdRaw['expiry']) {
			if (sessionIdRaw.expiry > new Date().value) {
				sessionId = sessionIdRaw.sessionId;
			}
		}
	} else {
		storeSessionId(sessionId);
	}

	// All preperatory stuff
	$( '.nav-authed' ).hide();

	$( '#login-form' ).submit(function() {
		console.log('login-submit.click');
		login();
		return false;
	});

	if (!authenticated()) {
		console.log('!auth');
		loadPage('login-splash');
		$( '#loginDialog' ).modal('show');
		$( "#login-username" ).focus();
	} else {
		$( 'LI.nav-authed' ).show();
		if (location.href.indexOf('#') > 1) {
			var hash = location.href.substr(location.href.indexOf('#') + 1);
			if (typeof hash !== 'undefined' && hash != '') {
			// Load hash
				loadPage( hash );
			}
		} else {
			loadPage( "main" );
		}
	}
});
