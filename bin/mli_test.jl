#!/usr/bin/env julia

module Test

using InteractiveUtils

#mli = Matrix{Float32}(4566, 5528)

function main()
    mli = Vector{Float32}(undef, 22831)
    n, avg = 0, Float64(0.0)
    
    open("/media/nas1/Dszekcso/ASC_PS_proc/SLC/20160912.mli", "r") do f
        for ii = 1:4185
            read!(f, mli)
            
            for jj in 1:4566
                mli[jj] = bswap(mli[jj])
                avg += Float64(mli[jj] * mli[jj])
                avg += Float64(mli[jj] * mli[jj])
                avg += Float64(mli[jj] * mli[jj])
                n += 1
            end
        end
    end
    println("Avg: ", avg / Float64(n))
    #println(typeof(mli[1:8]))
    #println("Read: $read; Mean: $(m / read)")
end


@time main()
@time main()
#@code_warntype main()
#@code_native main()

end

