angular.module('vger', ['ngAnimate', 'ui']).controller('tasks_ctrl',
	function($scope, $http) {

		function monitor(path, ondata, onclose, onerror) {
			var websocket = new WebSocket('ws://' +
				window.location.host + path);

			websocket.onopen = onOpen;
			websocket.onclose = onClose;
			websocket.onmessage = onMessage;
			websocket.onerror = onError;

			function onOpen(evt) {}

			function onClose(evt) {
				// $scope.push_alert('socket close');
				if (onclose) {
					onclose(evt);
				}
			}

			function onMessage(evt) {
				$scope.$apply(function() {
					ondata(JSON.parse(evt.data));
				});
			}

			function onError(evt) {
				// $scope.push_alert(evt.data);
				if (onerror) {
					onerror(evt.data);
				}
			}

			function doSend(message) {
				websocket.send(message);
			}

			return websocket;
		}
		$scope.tasks_max_size = 100000000;

		var allTasks = [];
		$scope.tasks = [];
		$scope.config = {
			'max-speed': '0'
		};
		$scope.task_filter = {Subscribe:'Single Tasks'}
		$http.get('/config').success(function(resp) {
			$scope.config = resp;
			var v = $scope.config['shutdown-after-finish'];
			$scope.config['shutdown-after-finish'] = (v == 'true');
		})

		function monitor_process() {
			monitor('/progress', function(data) {
				for (var i = data.length - 1; i >= 0; i--) {
					var item = data[i];
					item.StartDate = new Date(Date.parse(item.StartDate))
					if (item.Subscribe == '') {
						item.Subscribe = 'Single Tasks';
					}
				};

				var collection = {};
				for (var i = data.length - 1; i >= 0; i--) {
					var item = data[i]
					if (item.Status != "Deleted") {
						collection[item.Name] = item;
					}
				};
				var tasks = $scope.tasks;
				for (var i = tasks.length - 1; i >= 0; i--) {
					var task = tasks[i];
					var source = collection[task.Name];
					if (source) {
						angular.forEach(source, function(val, key) {
							task[key] = val;
						});
						delete collection[task.Name];
					} else {
						tasks.splice(i, 1);
					}
				}

				angular.forEach(collection, function(val, key) {
					tasks.push(val)
				});

				var subscribeMap = {};
				angular.forEach($scope.subscribes, function(val, key) {
					val.Badge = 0;
					subscribeMap[val.Name] = val;
				});

				angular.forEach(tasks, function(val) {
					if ((val.Status == 'Downloading') || (val.Status == 'Queued')
						|| (val.Status == 'Stopped')
						|| ((val.Status=='Finished')&&(subscribeMap[val.Subscribe].Duration==0))
						|| ((val.Status=='Finished')&&(val.LastPlaying<subscribeMap[val.Subscribe].Duration))) {
						subscribeMap[val.Subscribe].Badge++;
					}
				});
			}, monitor_process);
		}

		$http.get('/subscribe').success(function (subscribes) {
			angular.forEach(subscribes, function (s) {
				$scope.subscribes.push(s);
			});
			monitor_process();
		})

		$scope.new_url = document.getElementById('new-url').value;


		$scope.parse_duration = function(dur) {
			var sec = Math.floor(dur / 1000000000);
			var min = Math.floor(sec / 60);
			var hour = Math.floor(min / 60);
			var day = Math.floor(hour / 24);

			return (day > 0 ? day + 'd' : '') + (hour > 0 ? hour % 24 + 'h' : '') + (min > 0 ? min % 60 + 'm' : '') + (sec > 0 ? sec % 60 + 's' : '');
		}

		$scope.send_open = function(task) {
			if (task.Status == 'New') {
				return;
			}

			$http.get('/open/' + task.Name).success(function(resp) {
				resp && $scope.push_alert(resp);
			});
		}
		$scope.send_resume = function(task) {
			$http.get('/resume/' + task.Name).success(function(resp) {
				resp && $scope.push_alert(resp);
			});
		}
		$scope.send_stop = function(task) {
			$http.get('/stop/' + task.Name).success(function(resp) {
				resp && $scope.push_alert(resp);
			});
		}
		$scope.send_limit = function($event) {
			$http.get('/limit/' + $event.target.value).success(function(resp) {
				resp && $scope.push_alert(resp);
			});
		};
		$scope.send_simultaneous_downloads = function() {
			$http.post('/config/simultaneous', $scope.config['simultaneous-downloads'])
				.success(function () {});
		}
		$scope.send_play = function(task) {
			$http.get('/play/' + task.Name).success(function(resp) {
				resp && $scope.push_alert(resp);
			})
		};

		$scope.waiting = false;

		$scope.thunder_commit = {
			name: "",
			url: "",
			verifycode: ""
		};
		$scope.thunder_needverifycode = false;

		$scope.new_thunder_task = function() {
			$scope.waiting = true;

			var text = JSON.stringify($scope.thunder_commit);

			$http.post('/thunder/new', text).success(function(data) {
				$scope.waiting = false;
				if (typeof data == 'string') {
					if (data == 'Need verify code') {
						$scope.thunder_needverifycode = true;
						document.getElementById('verifycode').src='/thunder/verifycode/'+(new Date);
						return;
					} else {
						$scope.thunder_needverifycode = false;
					}

					$scope.push_alert(data);
					return;
				}

				$scope.thunder_needverifycode = false;

				for (var i = data.length - 1; i >= 0; i--) {
					var item = data[i];
					item.loading = false;

					var j = item.Name.lastIndexOf('\/');
					item.Name = item.Name.substring(j + 1);
				}
				if (data.length == 1 && data[0].Percent == 100) {
					$scope.waiting = true;
					if ($scope.thunder_commit.name) {
						data[0].Name = $scope.thunder_commit.name;
					}

					$scope.download_bt_files(data[0]);
				} else {
					$scope.bt_files = data;
				}
			});
		}

		function new_task(url) {
			$scope.waiting = true;
			if (url.indexOf('lixian.vip.xunlei.com') != -1 ||
				url.indexOf('youtube.com') != -1 ||
				/.*dmg|.*zip|.*rar|.*exe|.*iso|.*pkg|.*gz/.test(url)) {
				$http.post('/new/', url).success(function(resp) {
					if (!resp) {
						url = '';
					}
					$scope.waiting = false;
					resp && $scope.push_alert(resp);
				}).error(function() {
					$scope.waiting = false;
				});
			} else {
				$scope.thunder_commit.url = url;
				$scope.thunder_commit.name = "";
				$scope.new_thunder_task();
			}
		};

		$scope.subscribes = [{Name:"Single Tasks"}];
		$scope.new_subscribe = function (url) {
			$scope.waiting = true;
			$http.post('/subscribe/new', url).success(function (data) {
				if (typeof data == 'string') {
					$scope.push_alert(data);
					return;
				}
				$scope.waiting = false;
				$scope.task_filter.Subscribe = data.Name;
				angular.forEach($scope.subscribes, function(s) {
					if (s.Name == data.Name) {
						$scope.switch_subscribe(data);
						throw "subscribe exists";
					}
				});
				$scope.subscribes.push(data);

				$scope.switch_subscribe(data);

				$scope.new_url = '';
			});
		}
		$scope.currentSubscribe = {Name:'Single Tasks'};
		$scope.switch_subscribe = function(s) {
			$scope.task_filter.Subscribe = s.Name;

			var current = get_subscribe(s.Name);
			angular.forEach(current, function(v, k){
				$scope.currentSubscribe[k] = current[k];
			});

			$scope.tasks_max_size = 11;
			setTimeout(function() {
				$scope.$apply(function() {
					$scope.tasks_max_size = 10000000;
				});
				setTimeout(function() {
					var top = parseInt($($('#tasks-list .highlight-task')[0]).data('order'))*80;
					if (top==NaN) top = 0;
					// console.log(top);
					$('#tasks-list').scrollTop(top);
				}, 350);
			}, 500);
		}

		function get_subscribe(name) {
			for (var i = 0; i < $scope.subscribes.length; i++) {
				var s = $scope.subscribes[i];
				if (name == s.Name) {
					return s;
				}
			}

			return $scope.subscribes[0];
		}

		$scope.get_bt_file_status = function(percent) {
			return (percent == 100) ? 'Finished' : percent + '%'
		}

		$scope.download_bt_files = function(file) {
			file.loading = true;

			$http.post('/new/' + file.Name, file.DownloadURL).success(
				function(resp) {
					file.loading = false;
					$scope.waiting = false;
					if (resp) $scope.push_alert(resp);
					else {
						// $scope.bt_files = [];
						$scope.new_url = '';
					}
				}).error(function() {
				file.loading = false;
				$scope.waiting = false;
			});
		};
		$scope.bt_files = [];

		$scope.move_to_trash = function(task) {
			$http.get('/trash/' + task.Name).success(
				function(resp) {
					resp && $scope.push_alert(resp)
				});
		};

		$scope.set_autoshutdown = function() {
			$http.post('/autoshutdown', $scope.config['shutdown-after-finish']?'true':'false')
				.success(function() {});
		};


		//subtitles
		$scope.subtitles = [];
		$scope.subtitles_movie_name = '';

		$scope.ws_search_subtitles = null;

		$scope.search_subtitles = function(name) {
			if ($scope.ws_search_subtitles) {
				return;
			}

			$scope.subtitles = [];
			$scope.subtitles_movie_name = name;
			$scope.waiting = true;

			$scope.ws_search_subtitles = monitor('/subtitles/search/' + name, function(data) {
				if ($scope.ws_search_subtitles != null) {
					$scope.nosubtitles = false;
					data.loading = false;

					//truncate description
					data.FullDescription = data.Description;
					var description = data.Description;
					if (description.length > 73)
						data.Description = description.substr(0, 35) + '...' + description.substr(description.length - 35, 35);

					$scope.subtitles.push(data);
					$scope.waiting = true;
				}
			}, function() {
				// if ($scope.ws_search_subtitles != null) {
				// 	if ($scope.subtitles.length == 0) {
				// 		$scope.nosubtitles = true;
				// 	}
				// 	$scope.waiting = false;
				// 	$scope.ws_search_subtitles = null;
				// }
				$scope.waiting = false;
				// $scope.ws_search_subtitles = null;
			}, function() {
				if ($scope.ws_search_subtitles != null) {
					if ($scope.subtitles.length == 0) {
						$scope.nosubtitles = true;
					}
					$scope.waiting = false;
					$scope.ws_search_subtitles = null;
				}
			});
		};
		$scope.stop_search_subtitles = function() {
			if ($scope.ws_search_subtitles) {
				var ws = $scope.ws_search_subtitles;
				$scope.ws_search_subtitles = null;

				if ($scope.waiting == false) {
					ws.close();
				}

				$scope.waiting = false;
				$scope.nosubtitles = false;
				$scope.subtitles = [];
			}
		};

		$scope.download_subtitles = function(sub) {
			sub.loading = true;
			var input = JSON.stringify({'name':sub.Description, 'url':sub.URL});
			$http.post('/subtitles/download/' + $scope.subtitles_movie_name, input).success(function(resp) {
				sub.loading = false;
				$scope.stop_search_subtitles();
				if (resp) {
					$scope.push_alert(resp);
				}
			})
		};

		$scope.go = function() {
			$scope.waiting = true
			if (/www.yyets.com\/resource\/[0-9+]/.test($scope.new_url)) {
				$scope.new_subscribe($scope.new_url);
			} else if (/.+\:\/\/.+|^magnet\:\?.+/.test($scope.new_url)) {
				new_task($scope.new_url);
			} else {
				$scope.search_subtitles($scope.new_url)
			}
		};
		$scope.download_task = function (task) {
			// alert(task.Original);
			if (!task.Original) {
				$scope.push_alert("No original URL.")
				return;
			}
			$scope.thunder_commit.url = task.Original;
			$scope.thunder_commit.name = task.Name;
			$scope.new_thunder_task();
		};
		$scope.google_subtitles = function() {
			var name = $scope.subtitles_movie_name;
			name = name.replace(/(.*)[.](mkv|avi|mp4|rm|rmvb)/, '$1').replace(/(.*)-.*/, '$1') + ' subtitles';
			window.open("http://www.google.com/search?q=" + encodeURIComponent(name));
			$scope.nosubtitles = false;
		};
		$scope.addic7ed_subtitles = function() {
			var name = $scope.subtitles_movie_name;
			var i = name.lastIndexOf('.');
			if (i != -1) {
				name = name.substring(0, i);
			}

			i = name.lastIndexOf('-');
			if (i != -1) {
				name = name.substring(0, i);
			}

			name = name.replace(/720p|x[.]264|BluRay|DTS|x264|1080p|H[.]264|AC3|[.]ENG|[.]BD|[.]Rip|H264|HDTV|-IMMERSE|-DIMENSION|xvid|[[]PublicHD[]]|[.]Rus|Chi_Eng|DD5[.]1|HR-HDTV|[.]HDTVrip|[.]AAC|[0-9]+x[0-9]+|blu-ray|Remux|dxva|dvdscr|WEB-DL/ig, '');
			name = name.replace(/[\u4E00-\u9FFF]+/ig, '');
			name = name.replace(/[.]/g, ' ');

			window.open("http://www.addic7ed.com/search.php?search=" + encodeURIComponent(name));
			$scope.nosubtitles = false;
		}


		$scope.parse_time = function(time) {
			var d = new Date(time * 1000);
			return d.format('ddd mmm dd')
		}
		$scope.upload_torrent = function($event) {
			$event.preventDefault();

			if ($event.dataTransfer.files.length == 0) {
				var items = $event.dataTransfer.items;
				if (items && items.length > 0 && items[0].type == 'text/plain') {
					items[0].getAsString(function(str) {
						$scope.$apply(function() {
							$scope.new_url = str;
						})
					});
				}
				return;
			}

			if (!/[.]torrent$/.test($event.dataTransfer.files[0].name)) {
				$scope.push_alert('Only support .torrent file!')
				return;
			}

			var xhr = new XMLHttpRequest;
			var fd = new FormData();
			fd.append('torrent', $event.dataTransfer.files[0], 'torrent');

			$scope.waiting = true;

			xhr.open('POST', '/thunder/torrent');
			xhr.send(fd);
			xhr.onreadystatechange = function() {
				if (this.readyState == this.DONE) {
					// if (!$scope.waiting)) return;
					$scope.$apply(function() {
						$scope.waiting = false;
					});

					if (this.status == 200 && this.responseText != null) {
						var responseText = this.responseText;
						$scope.$apply(function() {
							if (responseText[0] != '[') {
								responseText && $scope.push_alert(responseText);
							} else {
								$scope.bt_files = JSON.parse(responseText);
							}
						});
					}
				}
			}
		}

		$scope.alerts = [];
		$scope.push_alert = function(content, title) {
			title = title || 'Error';
			$scope.alerts.push({
				'title': title,
				'content': content
			});
		}
		$scope.pop_alert = function() {
			$scope.alerts.pop();
		}

		window.onload = function() {
			setTimeout(function() {
				document.getElementById('box-overlay').style.display = '';
			}, 500);
			var ele = document.getElementById('new-url');
			ele.value = getCookie('input');
			ele.select();

			$scope.new_url = ele.value;
		}
	}

);


function getCookie(c_name) {
	var c_value = document.cookie;
	var c_start = c_value.indexOf(" " + c_name + "=");
	if (c_start == -1) {
		c_start = c_value.indexOf(c_name + "=");
	}
	if (c_start == -1) {
		c_value = null;
	} else {
		c_start = c_value.indexOf("=", c_start) + 1;
		var c_end = c_value.indexOf(";", c_start);
		if (c_end == -1) {
			c_end = c_value.length;
		}
		c_value = unescape(c_value.substring(c_start, c_end));
	}
	return c_value;
}

function setCookie(c_name, value, exdays) {
	var exdate = new Date();
	exdate.setDate(exdate.getDate() + exdays);
	var c_value = escape(value) + ((exdays == null) ? "" : "; expires=" + exdate.toUTCString());
	document.cookie = c_name + "=" + c_value;
}
window.onbeforeunload = function() {
	setCookie('input', document.getElementById('new-url').value, 10000)
}