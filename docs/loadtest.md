# DB Version: 16
# OS Type: linux
# DB Type: oltp
# Total Memory (RAM): 64 GB
# CPUs num: 16
# Data Storage: ssd

max_connections = 300
shared_buffers = 16GB
effective_cache_size = 48GB
maintenance_work_mem = 2GB
checkpoint_completion_target = 0.9
wal_buffers = 16MB
default_statistics_target = 100
random_page_cost = 1.1
effective_io_concurrency = 200
work_mem = 13981kB
huge_pages = try
min_wal_size = 2GB
max_wal_size = 8GB
max_worker_processes = 16
max_parallel_workers_per_gather = 4
max_parallel_workers = 16
max_parallel_maintenance_workers = 4
